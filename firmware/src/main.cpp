#include <M5Dial.h>
#include <cstdio>
#include <cstring>
#include <math.h>

static const char *kFw = "0.6.3";
static const uint32_t kHostTimeoutMs = 3000;
static const uint32_t kOverlayHoldMs = 2500;
static const int kDetentPulses = 4;

enum Screen { SCREEN_WAIT, SCREEN_IDLE, SCREEN_OVERLAY };

static Screen screen = SCREEN_WAIT;
static Screen drawn = (Screen)-1;
static uint32_t lastHostMs = 0;
static uint32_t lastEncMs = 0;
static long encPos = 0;
static String overlayBrand;
static String overlayTitle;
static String lineBuf;
static bool hostLink = false;
static bool dirty = true;
static const char *pending = nullptr;
static int rotDeg = 0;
static M5Canvas canvas(&M5Dial.Display);
static bool haveCanvas = false;
static int canvasDepth = 0;
static const uint32_t kI2CHz = 400000;
static const uint8_t kTofAddr = 0x29;
static const uint8_t kAccelAddr = 0x53;
static const uint8_t kMaxSlots = 64;
static const uint8_t kKindTof = 0;
static const uint8_t kKindAccel = 1;
struct SensorSlot {
  char id[20];
  uint8_t mux;
  uint8_t ch;
  uint8_t kind;
  bool seen;
  bool ok;
  bool reported;
  uint8_t failures;
};
static SensorSlot slots[kMaxSlots];
static uint8_t slotCount = 0;
static uint8_t muxAddrs[8];
static uint8_t muxCount = 0;
static uint32_t lastTofPollMs = 0;
static uint32_t lastAccelPollMs = 0;
static uint32_t lastSensorScanMs = 0;
static uint8_t nextInit = 0;
static bool scanRequested = true;

static m5::I2C_Class *bus = &M5.Ex_I2C;
static bool exReady = false;
static uint8_t i2cSda = 2;
static uint8_t i2cScl = 1;
static char i2cPort = 'b';

static bool i2cProbe(uint8_t address) {
  if (bus == &M5.Ex_I2C && !exReady) {
    return false;
  }
  return bus->scanID(address, kI2CHz);
}

static bool i2cWriteRaw(uint8_t addr, const uint8_t *data, size_t n) {
  if (!bus->start(addr, false, kI2CHz)) {
    return false;
  }
  bool ok = n == 0 || bus->write(data, n);
  return bus->stop() && ok;
}

static const char *kindName(uint8_t kind) { return kind == kKindAccel ? "accel" : "tof"; }

static void formatId(char *out, size_t n, uint8_t mux, uint8_t ch, uint8_t kind) {
  if (mux == 0) {
    snprintf(out, n, "root:%s", kindName(kind));
  } else {
    snprintf(out, n, "mux:%x:%u:%s", mux, ch, kindName(kind));
  }
}

static void muxIdle() {
  uint8_t zero = 0;
  for (uint8_t i = 0; i < muxCount; ++i) {
    i2cWriteRaw(muxAddrs[i], &zero, 1);
  }
}

static bool muxSelect(uint8_t mux, uint8_t channel) {
  muxIdle();
  if (mux == 0) {
    return true;
  }
  uint8_t mask = (uint8_t)(1U << channel);
  return i2cWriteRaw(mux, &mask, 1);
}

static int findSlot(uint8_t mux, uint8_t ch, uint8_t kind) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].mux == mux && slots[i].ch == ch && slots[i].kind == kind) {
      return i;
    }
  }
  return -1;
}

static int ensureSlot(uint8_t mux, uint8_t ch, uint8_t kind) {
  int idx = findSlot(mux, ch, kind);
  if (idx >= 0) {
    slots[idx].seen = true;
    return idx;
  }
  if (slotCount >= kMaxSlots) {
    return -1;
  }
  idx = slotCount++;
  formatId(slots[idx].id, sizeof(slots[idx].id), mux, ch, kind);
  slots[idx].mux = mux;
  slots[idx].ch = ch;
  slots[idx].kind = kind;
  slots[idx].seen = true;
  slots[idx].ok = false;
  slots[idx].reported = false;
  slots[idx].failures = 0;
  return idx;
}

static void reportSlot(uint8_t i) {
  if (i >= slotCount) {
    return;
  }
  SensorSlot &s = slots[i];
  Serial.printf(
      "{\"v\":1,\"t\":\"sensor\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"kind\":\"%s\",\"ok\":%s}\n",
      s.id, s.mux, s.ch, kindName(s.kind), s.ok ? "true" : "false");
  s.reported = true;
}

static void setSlotOk(int idx, bool ok) {
  if (idx < 0) {
    return;
  }
  SensorSlot &s = slots[idx];
  if (s.reported && s.ok == ok) {
    return;
  }
  s.ok = ok;
  if (!ok) {
    s.failures = 0;
  }
  reportSlot((uint8_t)idx);
}

static bool adxlWrite(uint8_t mux, uint8_t ch, uint8_t reg, uint8_t value) {
  if (!muxSelect(mux, ch)) {
    return false;
  }
  return bus->writeRegister8(kAccelAddr, reg, value, kI2CHz);
}

static bool adxlRead(uint8_t mux, uint8_t ch, uint8_t reg, uint8_t *data, size_t count) {
  if (!muxSelect(mux, ch)) {
    return false;
  }
  return bus->readRegister(kAccelAddr, reg, data, count, kI2CHz);
}

static bool initAccelSlot(int idx) {
  SensorSlot &s = slots[idx];
  uint8_t id = 0;
  if (!adxlRead(s.mux, s.ch, 0x00, &id, 1) || id != 0xE5) {
    return false;
  }
  return adxlWrite(s.mux, s.ch, 0x31, 0x08) && adxlWrite(s.mux, s.ch, 0x2C, 0x09) &&
         adxlWrite(s.mux, s.ch, 0x2D, 0x08);
}

static bool tofWr(uint16_t reg, const uint8_t *data, size_t n) {
  uint8_t buf[18];
  if (n + 2 > sizeof(buf)) {
    return false;
  }
  buf[0] = (uint8_t)(reg >> 8);
  buf[1] = (uint8_t)reg;
  memcpy(buf + 2, data, n);
  return i2cWriteRaw(kTofAddr, buf, n + 2);
}

static bool tofWr8(uint16_t reg, uint8_t v) { return tofWr(reg, &v, 1); }

static bool tofWr16(uint16_t reg, uint16_t v) {
  uint8_t d[2] = {(uint8_t)(v >> 8), (uint8_t)v};
  return tofWr(reg, d, 2);
}

static bool tofRd(uint16_t reg, uint8_t *data, size_t n) {
  uint8_t addr[2] = {(uint8_t)(reg >> 8), (uint8_t)reg};
  if (!bus->start(kTofAddr, false, kI2CHz) || !bus->write(addr, 2)) {
    bus->stop();
    return false;
  }
  if (!bus->restart(kTofAddr, true, kI2CHz)) {
    bus->stop();
    return false;
  }
  bool ok = bus->read(data, n, true);
  return bus->stop() && ok;
}

static bool tofRd8(uint16_t reg, uint8_t *v) { return tofRd(reg, v, 1); }

static bool tofRd16(uint16_t reg, uint16_t *v) {
  uint8_t d[2];
  if (!tofRd(reg, d, 2)) {
    return false;
  }
  *v = (uint16_t)((uint16_t)d[0] << 8 | d[1]);
  return true;
}

static const uint8_t kTofDefaultConfig[] = {
    0x12, 0x00, 0x00, 0x11, 0x02, 0x00, 0x02, 0x08, 0x00, 0x08, 0x10, 0x01, 0x01,
    0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x20, 0x0b, 0x00, 0x00, 0x02, 0x14, 0x21, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
    0x00, 0xc8, 0x00, 0x00, 0x38, 0xff, 0x01, 0x00, 0x08, 0x00, 0x00, 0x01, 0xcc,
    0x07, 0x01, 0xf1, 0x05, 0x00, 0xa0, 0x00, 0x80, 0x08, 0x38, 0x00, 0x00, 0x00,
    0x00, 0x0f, 0x89, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x07, 0x05,
    0x06, 0x06, 0x00, 0x00, 0x02, 0xc7, 0xff, 0x9B, 0x00, 0x00, 0x00, 0x01, 0x00,
    0x00};

static bool tofDataReady() {
  uint8_t mux = 0;
  uint8_t status = 0;
  if (!tofRd8(0x0030, &mux) || !tofRd8(0x0031, &status)) {
    return false;
  }
  uint8_t pol = ((mux & 0x10) >> 4) ? 0 : 1;
  return (status & 1) == pol;
}

static bool initTofSlot(int idx) {
  SensorSlot &s = slots[idx];
  if (!muxSelect(s.mux, s.ch) || !i2cProbe(kTofAddr)) {
    return false;
  }
  uint8_t boot = 0;
  for (int i = 0; i < 100; ++i) {
    if (tofRd8(0x00E5, &boot) && boot == 0x03) {
      break;
    }
    delay(10);
  }
  if (boot != 0x03) {
    return false;
  }
  uint16_t id = 0;
  if (!tofRd16(0x010F, &id) || id != 0xEBAA) {
    return false;
  }
  for (uint8_t addr = 0x2D; addr <= 0x87; ++addr) {
    if (!tofWr8(addr, kTofDefaultConfig[addr - 0x2D])) {
      return false;
    }
  }
  if (!tofWr8(0x0087, 0x40)) {
    return false;
  }
  for (int i = 0; i < 100 && !tofDataReady(); ++i) {
    delay(10);
  }
  tofWr8(0x0086, 0x01);
  tofWr8(0x0087, 0x00);
  tofWr8(0x0008, 0x09);
  tofWr8(0x000B, 0x00);
  tofWr16(0x0024, 0x0500);
  uint8_t zero4[4] = {0, 0, 0, 0};
  tofWr(0x006C, zero4, 4);
  return tofWr8(0x0087, 0x21);
}

static int countHits() {
  int n = 0;
  for (uint8_t addr = 0x70; addr <= 0x77; ++addr) {
    if (bus->scanID(addr, kI2CHz)) {
      n++;
    }
  }
  if (bus->scanID(kTofAddr, kI2CHz)) {
    n++;
  }
  if (bus->scanID(kAccelAddr, kI2CHz)) {
    n++;
  }
  return n;
}

static bool tryPins(int sda, int scl) {
  M5.Ex_I2C.release();
  if (!M5.Ex_I2C.begin((i2c_port_t)0, sda, scl)) {
    exReady = false;
    return false;
  }
  bus = &M5.Ex_I2C;
  exReady = true;
  i2cSda = (uint8_t)sda;
  i2cScl = (uint8_t)scl;
  i2cPort = (sda == 13 || sda == 15 || scl == 13 || scl == 15) ? 'a' : 'b';
  return true;
}

static void pickBus() {
  const int cand[][2] = {{2, 1}, {1, 2}, {13, 15}, {15, 13}};
  int bestN = -1;
  int bestSda = 2;
  int bestScl = 1;
  for (int i = 0; i < 4; ++i) {
    if (!tryPins(cand[i][0], cand[i][1])) {
      continue;
    }
    int n = countHits();
    if (n > bestN) {
      bestN = n;
      bestSda = cand[i][0];
      bestScl = cand[i][1];
    }
  }
  tryPins(bestSda, bestScl);
  Serial.printf("{\"v\":1,\"t\":\"i2c\",\"port\":\"%c\",\"sda\":%u,\"scl\":%u,\"n\":%d}\n",
                i2cPort, i2cSda, i2cScl, bestN < 0 ? 0 : bestN);
}

static void discoverMuxes() {
  pickBus();
  muxCount = 0;
  if (!exReady) {
    return;
  }
  for (uint8_t addr = 0x70; addr <= 0x77; ++addr) {
    if (!i2cProbe(addr)) {
      continue;
    }
    muxAddrs[muxCount++] = addr;
  }
}

static void markUnseenMissing() {
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (!slots[i].seen) {
      setSlotOk(i, false);
    }
  }
}

static void probeBus(uint8_t mux, uint8_t ch) {
  if (!muxSelect(mux, ch)) {
    return;
  }
  if (i2cProbe(0x29)) {
    ensureSlot(mux, ch, kKindTof);
  }
  if (i2cProbe(0x53)) {
    ensureSlot(mux, ch, kKindAccel);
  }
}

static void discoverSensors() {
  for (uint8_t i = 0; i < slotCount; ++i) {
    slots[i].seen = false;
  }
  discoverMuxes();
  for (uint8_t m = 0; m < muxCount; ++m) {
    for (uint8_t ch = 0; ch < 8; ++ch) {
      probeBus(muxAddrs[m], ch);
    }
  }
  muxIdle();
  probeBus(0, 0);
  markUnseenMissing();
  Serial.printf("{\"v\":1,\"t\":\"scan\",\"n\":%u}\n", slotCount);
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].seen && !slots[i].ok) {
      slots[i].reported = false;
    }
    reportSlot(i);
  }
}

static void initPending(uint32_t now) {
  if (now - lastSensorScanMs < 200) {
    return;
  }
  lastSensorScanMs = now;
  if (slotCount == 0) {
    return;
  }
  for (uint8_t n = 0; n < slotCount; ++n) {
    uint8_t i = nextInit;
    nextInit = (uint8_t)((nextInit + 1) % slotCount);
    if (!slots[i].seen || slots[i].ok) {
      continue;
    }
    bool ok = slots[i].kind == kKindAccel ? initAccelSlot(i) : initTofSlot(i);
    setSlotOk(i, ok);
    slots[i].failures = 0;
    return;
  }
}

static void scanSensors(uint32_t now) {
  if (scanRequested) {
    scanRequested = false;
    discoverSensors();
    lastSensorScanMs = now;
    return;
  }
  initPending(now);
}

static void sensorFailed(int idx) {
  if (idx < 0) {
    return;
  }
  if (++slots[idx].failures < 3) {
    return;
  }
  setSlotOk(idx, false);
}

static void pollTof(uint32_t now) {
  if (now - lastTofPollMs < 50) {
    return;
  }
  lastTofPollMs = now;
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].kind != kKindTof || !slots[i].ok) {
      continue;
    }
    if (!muxSelect(slots[i].mux, slots[i].ch)) {
      setSlotOk(i, false);
      continue;
    }
    delay(2);
    uint16_t mm = 0;
    if (!tofRd16(0x0096, &mm)) {
      sensorFailed(i);
      continue;
    }
    if (tofDataReady()) {
      tofWr8(0x0086, 0x01);
    }
    slots[i].failures = 0;
    Serial.printf("{\"v\":1,\"t\":\"tof\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"mm\":%u}\n",
                  slots[i].id, slots[i].mux, slots[i].ch, mm);
  }
}

static void pollAccel(uint32_t now) {
  if (now - lastAccelPollMs < 20) {
    return;
  }
  lastAccelPollMs = now;
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].kind != kKindAccel || !slots[i].ok) {
      continue;
    }
    uint8_t data[6];
    if (!adxlRead(slots[i].mux, slots[i].ch, 0x32, data, sizeof(data))) {
      sensorFailed(i);
      continue;
    }
    slots[i].failures = 0;
    int16_t rawX = (int16_t)((uint16_t)data[1] << 8 | data[0]);
    int16_t rawY = (int16_t)((uint16_t)data[3] << 8 | data[2]);
    int16_t rawZ = (int16_t)((uint16_t)data[5] << 8 | data[4]);
    Serial.printf(
        "{\"v\":1,\"t\":\"accel\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"x\":%d,\"y\":%d,\"z\":%d}\n",
        slots[i].id, slots[i].mux, slots[i].ch, rawX * 4, rawY * 4, rawZ * 4);
  }
}

static uint16_t rgb565(uint8_t r, uint8_t g, uint8_t b) {
  return ((r & 0xF8) << 8) | ((g & 0xFC) << 3) | (b >> 3);
}

static const uint16_t COL_BG = rgb565(15, 23, 42);
static const uint16_t COL_TEXT = rgb565(226, 232, 240);
static const uint16_t COL_MUTED = rgb565(148, 163, 184);
static const uint16_t COL_TILE = rgb565(56, 189, 248);
static const uint16_t COL_STACK = rgb565(52, 211, 153);
static const uint16_t COL_PANEL = rgb565(30, 41, 59);
static const uint16_t COL_OK = rgb565(250, 204, 21);
static const uint16_t COL_OK_BG = rgb565(69, 26, 3);

static bool jsonHasType(const String &line, const char *t) {
  String pat = String("\"t\":\"") + t + "\"";
  return line.indexOf(pat) >= 0;
}

static String jsonString(const String &line, const char *key) {
  String needle = String("\"") + key + "\":\"";
  int start = line.indexOf(needle);
  if (start < 0) {
    return "";
  }
  start += needle.length();
  String out;
  out.reserve(64);
  bool esc = false;
  for (int i = start; i < (int)line.length(); i++) {
    char c = line[i];
    if (esc) {
      out += c;
      esc = false;
      continue;
    }
    if (c == '\\') {
      esc = true;
      continue;
    }
    if (c == '"') {
      break;
    }
    out += c;
  }
  return out;
}

static void sendRaw(const char *s) { Serial.println(s); }

static void sendHello() {
  sendRaw("CLI-DIAL/1");
  Serial.printf("{\"v\":1,\"t\":\"hello\",\"fw\":\"%s\",\"dev\":\"cli-dial\"}\n", kFw);
}

static void sendEnc(int d) {
  Serial.printf("{\"v\":1,\"t\":\"enc\",\"d\":%d}\n", d);
}

static void sendTap(const char *id) {
  Serial.printf("{\"v\":1,\"t\":\"tap\",\"id\":\"%s\"}\n", id);
}

static void sendBtn() { sendRaw("{\"v\":1,\"t\":\"btn\",\"id\":\"a\"}"); }

static void sendPong() { sendRaw("{\"v\":1,\"t\":\"pong\"}"); }

static int jsonInt(const String &line, const char *key, int def) {
  String needle = String("\"") + key + "\":";
  int start = line.indexOf(needle);
  if (start < 0) {
    return def;
  }
  return line.substring(start + needle.length()).toInt();
}

static int wrapDeg(int deg) {
  deg %= 360;
  if (deg < 0) {
    deg += 360;
  }
  return deg;
}

static bool softRot() { return (rotDeg % 90) != 0; }

static void dropCanvas() {
  if (haveCanvas) {
    canvas.deleteSprite();
    haveCanvas = false;
  }
  canvasDepth = 0;
}

static bool ensureCanvas() {
  if (haveCanvas) {
    return true;
  }
  canvas.setPsram(false);
  canvas.setColorDepth(16);
  if (ESP.getFreeHeap() > 180000) {
    haveCanvas = canvas.createSprite(240, 240);
    if (haveCanvas) {
      canvasDepth = 16;
    }
  }
  if (!haveCanvas) {
    canvas.setColorDepth(8);
    haveCanvas = canvas.createSprite(240, 240);
    if (haveCanvas) {
      canvasDepth = 8;
    }
  }
  if (haveCanvas) {
    canvas.setPivot(120, 120);
  }
  return haveCanvas;
}

static void applyRotation(int deg) {
  deg = wrapDeg(deg);
  if (deg == rotDeg) {
    return;
  }
  rotDeg = deg;
  if (softRot()) {
    M5Dial.Display.setRotation(0);
    ensureCanvas();
  } else {
    dropCanvas();
    M5Dial.Display.setRotation(deg / 90);
  }
  dirty = true;
  drawn = (Screen)-1;
}

static void mapTouch(int px, int py, int *ox, int *oy) {
  if (!softRot()) {
    *ox = px;
    *oy = py;
    return;
  }
  float th = -rotDeg * 0.01745329252f;
  float c = cosf(th);
  float s = sinf(th);
  float dx = (float)(px - 120);
  float dy = (float)(py - 120);
  *ox = (int)lroundf(120.0f + dx * c - dy * s);
  *oy = (int)lroundf(120.0f + dx * s + dy * c);
}

static void click() { M5Dial.Speaker.tone(4000, 20); }

static bool inFace(int x, int y) {
  int dx = x - 120;
  int dy = y - 120;
  return dx * dx + dy * dy <= 120 * 120;
}

static bool inOK(int x, int y) {
  return pending && x >= 70 && x <= 170 && y >= 152 && y <= 210 && inFace(x, y);
}

static bool inTile(int x, int y) {
  return inFace(x, y) && x < 120 && y < 148;
}

static bool inStack(int x, int y) {
  return inFace(x, y) && x >= 120 && y < 148;
}

static LovyanGFX &face() {
  if (haveCanvas && softRot()) {
    return canvas;
  }
  return M5Dial.Display;
}

static void drawWaiting() {
  auto &d = face();
  d.fillScreen(COL_BG);
  d.setTextDatum(middle_center);
  d.setTextColor(COL_MUTED, COL_BG);
  d.setTextSize(2);
  d.drawString("Waiting", 120, 120);
}

static void drawIdle() {
  auto &d = face();
  d.fillScreen(COL_BG);
  int tileY = pending ? 42 : 70;
  int tileH = pending ? 86 : 100;
  uint16_t tileFill = COL_PANEL;
  uint16_t stackFill = COL_PANEL;
  uint16_t tileEdge = COL_TILE;
  uint16_t stackEdge = COL_STACK;
  if (pending && strcmp(pending, "tile") == 0) {
    tileFill = rgb565(8, 47, 73);
    tileEdge = COL_OK;
  }
  if (pending && strcmp(pending, "stack") == 0) {
    stackFill = rgb565(6, 78, 59);
    stackEdge = COL_OK;
  }
  d.fillRoundRect(18, tileY, 95, tileH, 16, tileFill);
  d.fillRoundRect(127, tileY, 95, tileH, 16, stackFill);
  d.drawRoundRect(18, tileY, 95, tileH, 16, tileEdge);
  d.drawRoundRect(127, tileY, 95, tileH, 16, stackEdge);
  d.setTextDatum(middle_center);
  d.setTextSize(2);
  d.setTextColor(COL_TILE, tileFill);
  d.drawString("TILE", 65, tileY + tileH / 2);
  d.setTextColor(COL_STACK, stackFill);
  d.drawString("STACK", 174, tileY + tileH / 2);
  if (pending) {
    d.fillRoundRect(70, 158, 100, 46, 14, COL_OK_BG);
    d.drawRoundRect(70, 158, 100, 46, 14, COL_OK);
    d.setTextColor(COL_OK, COL_OK_BG);
    d.drawString("OK", 120, 181);
  }
}

static void drawOverlay() {
  auto &d = face();
  d.fillScreen(COL_BG);
  d.setTextDatum(middle_center);
  d.setTextSize(2);
  d.setTextColor(COL_TILE, COL_BG);
  String brand = overlayBrand.length() ? overlayBrand : "-";
  d.drawString(brand, 120, 88);
  d.setTextSize(1);
  d.setTextColor(COL_TEXT, COL_BG);
  String title = overlayTitle;
  if (!title.length()) {
    title = "No matching CLI windows";
  }
  if (title.length() > 28) {
    title = title.substring(0, 27) + ".";
  }
  d.drawString(title, 120, 140);
}

static void paint() {
  if (!dirty && screen == drawn) {
    return;
  }
  switch (screen) {
  case SCREEN_WAIT:
    drawWaiting();
    break;
  case SCREEN_IDLE:
    drawIdle();
    break;
  case SCREEN_OVERLAY:
    drawOverlay();
    break;
  }
  drawn = screen;
  dirty = false;
  if (haveCanvas && softRot()) {
    canvas.setPivot(120, 120);
    if (canvasDepth == 16) {
      M5Dial.Display.pushImageRotateZoom(
          120, 120, 120, 120, (float)rotDeg, 1.0f, 1.0f, 240, 240,
          static_cast<const uint16_t *>(canvas.getBuffer()));
    } else {
      M5Dial.Display.pushImageRotateZoom(
          120, 120, 120, 120, (float)rotDeg, 1.0f, 1.0f, 240, 240,
          static_cast<const uint8_t *>(canvas.getBuffer()));
    }
  }
}

static void handleHostLine(const String &line) {
  String s = line;
  s.trim();
  if (!s.length()) {
    return;
  }
  lastHostMs = millis();
  bool wasLink = hostLink;
  hostLink = true;
  if (jsonHasType(s, "ping")) {
    sendPong();
  }
  if (jsonHasType(s, "state")) {
    overlayBrand = jsonString(s, "brand");
    overlayTitle = jsonString(s, "title");
    applyRotation(jsonInt(s, "rot", 0));
    if (screen == SCREEN_OVERLAY) {
      dirty = true;
    }
  }
  if (jsonHasType(s, "scan")) {
    scanRequested = true;
  }
  if (!wasLink) {
    dirty = true;
    scanRequested = true;
  }
}

static void readSerial() {
  while (Serial.available()) {
    char c = (char)Serial.read();
    if (c == '\r') {
      continue;
    }
    if (c == '\n') {
      handleHostLine(lineBuf);
      lineBuf = "";
      continue;
    }
    if (lineBuf.length() < 511) {
      lineBuf += c;
    }
  }
}

void setup() {
  auto cfg = M5.config();
  M5Dial.begin(cfg, true, false);
  Serial.begin(115200);
  unsigned long t0 = millis();
  while (!Serial && millis() - t0 < 2000) {
    delay(10);
  }
  M5Dial.Display.setRotation(0);
  encPos = M5Dial.Encoder.read();
  sendHello();
  lastHostMs = 0;
  hostLink = false;
  pending = nullptr;
  screen = SCREEN_WAIT;
  dirty = true;
  paint();
  scanRequested = true;
}

void loop() {
  M5Dial.update();
  readSerial();
  uint32_t now = millis();
  scanSensors(now);
  pollTof(now);
  pollAccel(now);

  if (hostLink && (now - lastHostMs) > kHostTimeoutMs) {
    hostLink = false;
    pending = nullptr;
    screen = SCREEN_WAIT;
    dirty = true;
  }

  long pos = M5Dial.Encoder.read();
  long delta = pos - encPos;
  if (delta <= -kDetentPulses || delta >= kDetentPulses) {
    int steps = (int)(delta / kDetentPulses);
    encPos += (long)steps * kDetentPulses;
    if (hostLink) {
      click();
      sendEnc(steps);
      lastEncMs = now;
      screen = SCREEN_OVERLAY;
      dirty = true;
    }
  }

  if (M5Dial.BtnA.wasClicked() && hostLink) {
    click();
    if (pending && screen != SCREEN_OVERLAY) {
      sendTap(pending);
      pending = nullptr;
      screen = SCREEN_IDLE;
    } else {
      sendBtn();
      screen = SCREEN_IDLE;
    }
    dirty = true;
  }

  auto t = M5Dial.Touch.getDetail();
  if (t.wasClicked() && hostLink && screen != SCREEN_OVERLAY) {
    int lx = t.x;
    int ly = t.y;
    if (softRot()) {
      mapTouch(t.x, t.y, &lx, &ly);
    }
    if (inOK(lx, ly)) {
      click();
      sendTap(pending);
      pending = nullptr;
      dirty = true;
    } else if (inTile(lx, ly)) {
      click();
      pending = "tile";
      dirty = true;
    } else if (inStack(lx, ly)) {
      click();
      pending = "stack";
      dirty = true;
    }
  }

  if (hostLink) {
    if (screen == SCREEN_OVERLAY && (now - lastEncMs) > kOverlayHoldMs) {
      screen = SCREEN_IDLE;
      dirty = true;
    } else if (screen == SCREEN_WAIT) {
      screen = SCREEN_IDLE;
      dirty = true;
    }
  }

  paint();
  delay(8);
}

#include <M5Dial.h>
#include <Wire.h>
#include <vl53l4cd_class.h>
#include <Adafruit_ADXL345_U.h>
#include <cstring>
#include <math.h>

static const char *kFw = "0.5.0";
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
static TwoWire sensorWire(1);
static VL53L4CD tof0(&sensorWire, -1);
static VL53L4CD tof1(&sensorWire, -1);
static VL53L4CD tof2(&sensorWire, -1);
static VL53L4CD tof3(&sensorWire, -1);
static VL53L4CD *tofSensors[4] = {&tof0, &tof1, &tof2, &tof3};
static bool sensorOK[5] = {false, false, false, false, false};
static bool sensorReported[5] = {false, false, false, false, false};
static uint8_t sensorFailures[5] = {0, 0, 0, 0, 0};
static uint32_t lastTofPollMs = 0;
static uint32_t lastAccelPollMs = 0;
static uint32_t lastSensorScanMs = 0;
static uint8_t nextSensorScan = 0;
static bool muxOK = false;

static bool muxSelect(uint8_t channel) {
  sensorWire.beginTransmission(0x70);
  sensorWire.write((uint8_t)(1U << channel));
  return sensorWire.endTransmission() == 0;
}

static bool i2cProbe(uint8_t address) {
  sensorWire.beginTransmission(address);
  return sensorWire.endTransmission() == 0;
}

static void reportSensor(uint8_t channel, const char *kind, bool ok) {
  if (sensorReported[channel] && sensorOK[channel] == ok) {
    return;
  }
  sensorOK[channel] = ok;
  sensorReported[channel] = true;
  Serial.printf("{\"v\":1,\"t\":\"sensor\",\"ch\":%u,\"kind\":\"%s\",\"ok\":%s}\n",
                channel, kind, ok ? "true" : "false");
}

static bool adxlWrite(uint8_t reg, uint8_t value) {
  if (!muxSelect(4)) return false;
  sensorWire.beginTransmission(0x53);
  sensorWire.write(reg);
  sensorWire.write(value);
  return sensorWire.endTransmission() == 0;
}

static bool adxlRead(uint8_t reg, uint8_t *data, size_t count) {
  if (!muxSelect(4)) return false;
  sensorWire.beginTransmission(0x53);
  sensorWire.write(reg);
  if (sensorWire.endTransmission(false) != 0) return false;
  if (sensorWire.requestFrom((uint8_t)0x53, count) != count) return false;
  for (size_t i = 0; i < count; ++i) data[i] = sensorWire.read();
  return true;
}

static bool initAccel() {
  uint8_t id = 0;
  if (!adxlRead(0x00, &id, 1) || id != 0xE5) return false;
  return adxlWrite(0x31, 0x08) && adxlWrite(0x2C, 0x09) && adxlWrite(0x2D, 0x08);
}

static bool initTof(uint8_t channel) {
  if (!muxSelect(channel) || !i2cProbe(0x29)) return false;
  VL53L4CD *sensor = tofSensors[channel];
  sensor->begin();
  if (sensor->InitSensor() != 0) return false;
  if (sensor->VL53L4CD_SetRangeTiming(50, 0) != 0) return false;
  return sensor->VL53L4CD_StartRanging() == 0;
}

static void markMuxMissing() {
  muxOK = false;
  for (uint8_t ch = 0; ch < 5; ++ch) {
    reportSensor(ch, ch == 4 ? "accel" : "tof", false);
    sensorFailures[ch] = 0;
  }
}

static void scanSensors(uint32_t now) {
  if (now - lastSensorScanMs < 400) return;
  lastSensorScanMs = now;
  if (!muxOK) {
    sensorWire.beginTransmission(0x70);
    if (sensorWire.endTransmission() != 0) {
      markMuxMissing();
      return;
    }
    muxOK = true;
  }
  uint8_t ch = nextSensorScan;
  nextSensorScan = (uint8_t)((nextSensorScan + 1) % 5);
  if (sensorOK[ch]) return;
  bool ok = ch == 4 ? initAccel() : initTof(ch);
  reportSensor(ch, ch == 4 ? "accel" : "tof", ok);
  sensorFailures[ch] = 0;
}

static void sensorFailed(uint8_t channel, const char *kind) {
  if (++sensorFailures[channel] < 3) return;
  sensorFailures[channel] = 0;
  reportSensor(channel, kind, false);
}

static void pollTof(uint32_t now) {
  if (now - lastTofPollMs < 50) return;
  lastTofPollMs = now;
  for (uint8_t ch = 0; ch < 4; ++ch) {
    if (!sensorOK[ch]) continue;
    if (!muxSelect(ch)) { markMuxMissing(); return; }
    uint8_t ready = 0;
    if (tofSensors[ch]->VL53L4CD_CheckForDataReady(&ready) != 0) {
      sensorFailed(ch, "tof");
      continue;
    }
    if (!ready) continue;
    VL53L4CD_Result_t result;
    if (tofSensors[ch]->VL53L4CD_GetResult(&result) != 0 ||
        tofSensors[ch]->VL53L4CD_ClearInterrupt() != 0) {
      sensorFailed(ch, "tof");
      continue;
    }
    sensorFailures[ch] = 0;
    if (result.range_status == 0 && result.distance_mm > 0) {
      Serial.printf("{\"v\":1,\"t\":\"tof\",\"ch\":%u,\"mm\":%u}\n", ch, result.distance_mm);
    }
  }
}

static void pollAccel(uint32_t now) {
  if (!sensorOK[4] || now - lastAccelPollMs < 20) return;
  lastAccelPollMs = now;
  uint8_t data[6];
  if (!adxlRead(0x32, data, sizeof(data))) { sensorFailed(4, "accel"); return; }
  sensorFailures[4] = 0;
  int16_t rawX = (int16_t)((uint16_t)data[1] << 8 | data[0]);
  int16_t rawY = (int16_t)((uint16_t)data[3] << 8 | data[2]);
  int16_t rawZ = (int16_t)((uint16_t)data[5] << 8 | data[4]);
  Serial.printf("{\"v\":1,\"t\":\"accel\",\"ch\":4,\"x\":%d,\"y\":%d,\"z\":%d}\n",
                rawX * 4, rawY * 4, rawZ * 4);
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
  if (!wasLink) {
    dirty = true;
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
  sensorWire.begin(13, 15, 400000);
  sensorWire.setTimeOut(20);
  encPos = M5Dial.Encoder.read();
  sendHello();
  lastHostMs = 0;
  hostLink = false;
  pending = nullptr;
  screen = SCREEN_WAIT;
  dirty = true;
  paint();
  for (uint8_t ch = 0; ch < 5; ++ch) {
    reportSensor(ch, ch == 4 ? "accel" : "tof", false);
  }
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

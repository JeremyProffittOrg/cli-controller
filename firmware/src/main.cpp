#include <M5Dial.h>
#include <cstring>

static const char *kFw = "0.2.0";
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

static void drawWaiting() {
  M5Dial.Display.fillScreen(COL_BG);
  M5Dial.Display.setTextDatum(middle_center);
  M5Dial.Display.setTextColor(COL_MUTED, COL_BG);
  M5Dial.Display.setTextSize(2);
  M5Dial.Display.drawString("Waiting", 120, 120);
}

static void drawIdle() {
  M5Dial.Display.fillScreen(COL_BG);
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
  M5Dial.Display.fillRoundRect(18, tileY, 95, tileH, 16, tileFill);
  M5Dial.Display.fillRoundRect(127, tileY, 95, tileH, 16, stackFill);
  M5Dial.Display.drawRoundRect(18, tileY, 95, tileH, 16, tileEdge);
  M5Dial.Display.drawRoundRect(127, tileY, 95, tileH, 16, stackEdge);
  M5Dial.Display.setTextDatum(middle_center);
  M5Dial.Display.setTextSize(2);
  M5Dial.Display.setTextColor(COL_TILE, tileFill);
  M5Dial.Display.drawString("TILE", 65, tileY + tileH / 2);
  M5Dial.Display.setTextColor(COL_STACK, stackFill);
  M5Dial.Display.drawString("STACK", 174, tileY + tileH / 2);
  if (pending) {
    M5Dial.Display.fillRoundRect(70, 158, 100, 46, 14, COL_OK_BG);
    M5Dial.Display.drawRoundRect(70, 158, 100, 46, 14, COL_OK);
    M5Dial.Display.setTextColor(COL_OK, COL_OK_BG);
    M5Dial.Display.drawString("OK", 120, 181);
  }
}

static void drawOverlay() {
  M5Dial.Display.fillScreen(COL_BG);
  M5Dial.Display.setTextDatum(middle_center);
  M5Dial.Display.setTextSize(2);
  M5Dial.Display.setTextColor(COL_TILE, COL_BG);
  String brand = overlayBrand.length() ? overlayBrand : "-";
  M5Dial.Display.drawString(brand, 120, 88);
  M5Dial.Display.setTextSize(1);
  M5Dial.Display.setTextColor(COL_TEXT, COL_BG);
  String title = overlayTitle;
  if (!title.length()) {
    title = "No matching CLI windows";
  }
  if (title.length() > 28) {
    title = title.substring(0, 27) + ".";
  }
  M5Dial.Display.drawString(title, 120, 140);
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
  encPos = M5Dial.Encoder.read();
  sendHello();
  lastHostMs = 0;
  hostLink = false;
  pending = nullptr;
  screen = SCREEN_WAIT;
  dirty = true;
  paint();
}

void loop() {
  M5Dial.update();
  readSerial();
  uint32_t now = millis();

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
    if (inOK(t.x, t.y)) {
      click();
      sendTap(pending);
      pending = nullptr;
      dirty = true;
    } else if (inTile(t.x, t.y)) {
      click();
      pending = "tile";
      dirty = true;
    } else if (inStack(t.x, t.y)) {
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

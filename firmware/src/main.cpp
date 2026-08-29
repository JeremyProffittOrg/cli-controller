#include <M5Dial.h>

static const char *kFw = "0.1.0";
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

static uint16_t rgb565(uint8_t r, uint8_t g, uint8_t b) {
  return ((r & 0xF8) << 8) | ((g & 0xFC) << 3) | (b >> 3);
}

static const uint16_t COL_BG = rgb565(15, 23, 42);
static const uint16_t COL_TEXT = rgb565(226, 232, 240);
static const uint16_t COL_MUTED = rgb565(148, 163, 184);
static const uint16_t COL_TILE = rgb565(56, 189, 248);
static const uint16_t COL_STACK = rgb565(52, 211, 153);
static const uint16_t COL_PANEL = rgb565(30, 41, 59);

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

static void drawWaiting() {
  M5Dial.Display.fillScreen(COL_BG);
  M5Dial.Display.setTextDatum(middle_center);
  M5Dial.Display.setTextColor(COL_MUTED, COL_BG);
  M5Dial.Display.setTextSize(2);
  M5Dial.Display.drawString("Waiting", 120, 120);
}

static void drawIdle() {
  M5Dial.Display.fillScreen(COL_BG);
  M5Dial.Display.fillRoundRect(18, 70, 95, 100, 16, COL_PANEL);
  M5Dial.Display.fillRoundRect(127, 70, 95, 100, 16, COL_PANEL);
  M5Dial.Display.drawRoundRect(18, 70, 95, 100, 16, COL_TILE);
  M5Dial.Display.drawRoundRect(127, 70, 95, 100, 16, COL_STACK);
  M5Dial.Display.setTextDatum(middle_center);
  M5Dial.Display.setTextSize(2);
  M5Dial.Display.setTextColor(COL_TILE, COL_PANEL);
  M5Dial.Display.drawString("TILE", 65, 120);
  M5Dial.Display.setTextColor(COL_STACK, COL_PANEL);
  M5Dial.Display.drawString("STACK", 174, 120);
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

static bool inFace(int x, int y) {
  int dx = x - 120;
  int dy = y - 120;
  return dx * dx + dy * dy <= 120 * 120;
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
    sendBtn();
    screen = SCREEN_IDLE;
    dirty = true;
  }

  auto t = M5Dial.Touch.getDetail();
  if (t.wasClicked() && hostLink && inFace(t.x, t.y)) {
    if (t.x < 120) {
      click();
      sendTap("tile");
    } else {
      click();
      sendTap("stack");
    }
    screen = SCREEN_IDLE;
    dirty = true;
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

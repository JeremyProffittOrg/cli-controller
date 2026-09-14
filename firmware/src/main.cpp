#include <M5Dial.h>
#include <Wire.h>
#include <Preferences.h>
#include <cstdio>
#include <cstring>
#include <math.h>

static const char *kFw = "0.7.2";
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
// ---------------------------------------------------------------------------
// I2C sensor layer
//
// Every I2C transaction is bounded: no driver loop spins until a device
// answers. VL53L4CD access is an in-tree register driver (ST ULD register
// map, BSD-3-Clause, STMicroelectronics) because the STM32duino wrapper's
// I2CRead loops forever when a sensor is unplugged mid-read.
//
// All time-of-flight sensors range concurrently, each behind its own mux
// channel; the loop polls data-ready without blocking and only reads a
// result when one is waiting. ADXL345 samples are gated on DATA_READY so the
// host never sees duplicated frames.
// ---------------------------------------------------------------------------

static const uint32_t kI2CHz = 400000;
static const uint16_t kI2CTimeoutMs = 25;
static const uint8_t kTofAddr = 0x29;
static const uint8_t kTofRootAlt = 0x2A;
static const uint8_t kAccelAddr = 0x53;
static const uint8_t kAccelAddrAlt = 0x1D;
static const uint8_t kAccelFormat = 0x0B;  // FULL_RES, +/-16 g, 3.9 mg/LSB
static const uint8_t kAccelRate = 0x09;    // 50 Hz output data rate
static const uint32_t kTofBudgetMs = 50;   // continuous ranging, 20 Hz per sensor
static const uint16_t kTofSignalKcps = 50;
static const uint16_t kTofSigmaMm = 120;
static const uint32_t kStallMs = 600;
static const uint32_t kTofBootMs = 300;
static const uint32_t kTofVhvMs = 500;
static const uint32_t kInitRetryMs = 100;
static const uint32_t kBgProbeMs = 40;
static const uint32_t kEmptyRescanMs = 3000;
static const uint8_t kFailLimit = 3;
static const uint8_t kBusErrLimit = 6;
static const uint8_t kMaxSlots = 64;
static const uint8_t kKindTof = 0;
static const uint8_t kKindAccel = 1;

struct SensorSlot {
  char id[24];
  uint8_t mux;
  uint8_t ch;
  uint8_t kind;
  uint8_t addr;
  bool seen;
  bool ok;
  bool reported;
  bool ranging;
  bool dropFirst;
  uint8_t intPol;
  uint8_t failures;
  uint32_t lastDataMs;
  uint32_t nextPollMs;
  uint16_t chip;      // model id: VL53L4CD 0xEBAA, ADXL345 0xE5
  int16_t offsetMm;   // VL53L4CD range offset applied at init (NVS "tofcal")
  bool calActive;
  uint8_t calCount;
  uint32_t calSum;
  uint16_t calTarget;
  uint32_t calStartMs;
  uint16_t errors;    // failed transactions since boot
  uint16_t reinits;   // successful inits after the first
  bool everOk;
};

static Preferences prefs;
static const uint8_t kCalSamples = 20;
static const uint32_t kCalTimeoutMs = 5000;

static SensorSlot slots[kMaxSlots];
static uint8_t slotCount = 0;
static uint8_t muxAddrs[8];
static uint8_t muxCount = 0;
static bool muxCacheValid = false;
static uint8_t curMux = 0;
static uint8_t curCh = 0xFF;
static uint32_t lastInitMs = 0;
static uint32_t lastBgProbeMs = 0;
static uint32_t lastEmptyRescanMs = 0;
static uint8_t nextInit = 0;
static uint8_t bgCursor = 0;
static bool scanRequested = true;
static bool rootTofFixed = false;    // root 0x29 device cannot move: skip 0x29 behind muxes
static bool rootAccelFixed = false;  // root 0x53 present: skip 0x53 behind muxes
static bool rootAccelAltFixed = false;
static uint8_t busErrors = 0;
static uint32_t busRecoveries = 0;

static bool exReady = false;
static uint8_t i2cSda = 2;
static uint8_t i2cScl = 1;
static char i2cPort = 'b';

static void sendLog(const char *msg) {
  Serial.printf("{\"v\":1,\"t\":\"log\",\"msg\":\"%s\"}\n", msg);
}

// ---- raw I2C -------------------------------------------------------------

static bool i2cWrite(uint8_t addr, const uint8_t *data, size_t n) {
  if (!exReady) {
    return false;
  }
  Wire.beginTransmission(addr);
  if (n > 0 && Wire.write(data, n) != n) {
    Wire.endTransmission(true);
    return false;
  }
  return Wire.endTransmission(true) == 0;
}

static bool i2cWriteRead(uint8_t addr, const uint8_t *wr, size_t wn, uint8_t *rd, size_t rn) {
  if (!exReady) {
    return false;
  }
  Wire.beginTransmission(addr);
  if (Wire.write(wr, wn) != wn) {
    Wire.endTransmission(true);
    return false;
  }
  if (Wire.endTransmission(false) != 0) {
    return false;
  }
  if (Wire.requestFrom((uint8_t)addr, (uint8_t)rn) != rn) {
    while (Wire.available()) {
      Wire.read();
    }
    return false;
  }
  for (size_t i = 0; i < rn; ++i) {
    rd[i] = (uint8_t)Wire.read();
  }
  return true;
}

static bool i2cRead(uint8_t addr, uint8_t *rd, size_t rn) {
  if (!exReady) {
    return false;
  }
  if (Wire.requestFrom((uint8_t)addr, (uint8_t)rn) != rn) {
    while (Wire.available()) {
      Wire.read();
    }
    return false;
  }
  for (size_t i = 0; i < rn; ++i) {
    rd[i] = (uint8_t)Wire.read();
  }
  return true;
}

static bool i2cProbe(uint8_t addr) { return i2cWrite(addr, nullptr, 0); }

static bool i2cBegin(int sda, int scl) {
  Wire.end();
  if (!Wire.begin(sda, scl, kI2CHz)) {
    exReady = false;
    return false;
  }
  Wire.setTimeOut(kI2CTimeoutMs);
  exReady = true;
  muxCacheValid = false;
  return true;
}

// Bit-bang bus recovery: a slave left mid-transfer holds SDA low until it
// sees enough clocks to finish the byte. Clock SCL until SDA releases, then
// issue a STOP, then re-open the driver.
static void i2cRecover() {
  busRecoveries++;
  busErrors = 0;
  Wire.end();
  pinMode(i2cSda, OUTPUT_OPEN_DRAIN);
  pinMode(i2cScl, OUTPUT_OPEN_DRAIN);
  digitalWrite(i2cSda, HIGH);
  digitalWrite(i2cScl, HIGH);
  delayMicroseconds(10);
  for (int i = 0; i < 9 && digitalRead(i2cSda) == LOW; ++i) {
    digitalWrite(i2cScl, LOW);
    delayMicroseconds(5);
    digitalWrite(i2cScl, HIGH);
    delayMicroseconds(5);
  }
  digitalWrite(i2cSda, LOW);
  delayMicroseconds(5);
  digitalWrite(i2cScl, HIGH);
  delayMicroseconds(5);
  digitalWrite(i2cSda, HIGH);
  delayMicroseconds(10);
  i2cBegin(i2cSda, i2cScl);
  char msg[48];
  snprintf(msg, sizeof(msg), "i2c bus recovery %lu", (unsigned long)busRecoveries);
  sendLog(msg);
}

// Called when a transaction to a device that was known-good fails.
static void noteBusError() {
  if (++busErrors >= kBusErrLimit) {
    i2cRecover();
  }
}

static void noteBusOk() { busErrors = 0; }

// ---- PCA9548 mux ----------------------------------------------------------

static bool muxWrite(uint8_t mux, uint8_t mask) {
  if (!i2cWrite(mux, &mask, 1)) {
    return false;
  }
  uint8_t rb = 0xFF;
  if (!i2cRead(mux, &rb, 1)) {
    return false;
  }
  return rb == mask;
}

static void muxCloseAll() {
  for (uint8_t i = 0; i < muxCount; ++i) {
    muxWrite(muxAddrs[i], 0);
  }
  curMux = 0;
  curCh = 0xFF;
  muxCacheValid = true;
}

static bool muxSelect(uint8_t mux, uint8_t ch) {
  if (!exReady) {
    return false;
  }
  if (!muxCacheValid) {
    muxCloseAll();
  }
  if (mux == 0) {
    if (curMux != 0) {
      if (!muxWrite(curMux, 0)) {
        muxCacheValid = false;
        return false;
      }
      curMux = 0;
      curCh = 0xFF;
    }
    return true;
  }
  if (curMux == mux && curCh == ch) {
    return true;
  }
  if (curMux != 0 && curMux != mux) {
    if (!muxWrite(curMux, 0)) {
      muxCacheValid = false;
      return false;
    }
    curMux = 0;
    curCh = 0xFF;
  }
  if (!muxWrite(mux, (uint8_t)(1U << ch))) {
    muxCacheValid = false;
    return false;
  }
  curMux = mux;
  curCh = ch;
  return true;
}

// ---- VL53L4CD register driver --------------------------------------------

static const uint16_t kTofRegSlaveAddr = 0x0001;
static const uint16_t kTofRegOscFreq = 0x0006;
static const uint16_t kTofRegVhvLoopBound = 0x0008;
static const uint16_t kTofRegGpioHvMux = 0x0030;
static const uint16_t kTofRegGpioTioStatus = 0x0031;
static const uint16_t kTofRegRangeConfigA = 0x005E;
static const uint16_t kTofRegRangeConfigB = 0x0061;
static const uint16_t kTofRegSigmaThresh = 0x0064;
static const uint16_t kTofRegMinCountRate = 0x0066;
static const uint16_t kTofRegInterMeasurement = 0x006C;
static const uint16_t kTofRegIntClear = 0x0086;
static const uint16_t kTofRegSystemStart = 0x0087;
static const uint16_t kTofRegResultBlock = 0x0089;
static const uint16_t kTofRegOscCalibrate = 0x00DE;
static const uint16_t kTofRegFirmwareStatus = 0x00E5;
static const uint16_t kTofRegModelId = 0x010F;
static const uint16_t kTofModelId = 0xEBAA;

// ST default configuration for registers 0x2D..0x87 (VL53L4CD ULD 1.0).
static const uint8_t kTofDefaultConfig[91] = {
    0x12, 0x00, 0x00, 0x11, 0x02, 0x00, 0x02, 0x08, 0x00, 0x08, 0x10, 0x01, 0x01,
    0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20,
    0x0b, 0x00, 0x00, 0x02, 0x14, 0x21, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00,
    0xc8, 0x00, 0x00, 0x38, 0xff, 0x01, 0x00, 0x08, 0x00, 0x00, 0x01, 0xcc, 0x07,
    0x01, 0xf1, 0x05, 0x00, 0xa0, 0x00, 0x80, 0x08, 0x38, 0x00, 0x00, 0x00, 0x00,
    0x0f, 0x89, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x07, 0x05, 0x06,
    0x06, 0x00, 0x00, 0x02, 0xc7, 0xff, 0x9B, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00};

struct TofResult {
  uint8_t status;
  uint16_t mm;
  uint16_t sigmaMm;
  uint16_t sigPerSpad;
  uint16_t ambPerSpad;
  uint16_t spad;
};

static bool tofRd(uint8_t a, uint16_t reg, uint8_t *buf, size_t n) {
  uint8_t r[2] = {(uint8_t)(reg >> 8), (uint8_t)(reg & 0xFF)};
  return i2cWriteRead(a, r, 2, buf, n);
}

static bool tofRd8(uint8_t a, uint16_t reg, uint8_t *v) { return tofRd(a, reg, v, 1); }

static bool tofRd16(uint8_t a, uint16_t reg, uint16_t *v) {
  uint8_t b[2];
  if (!tofRd(a, reg, b, 2)) {
    return false;
  }
  *v = (uint16_t)(((uint16_t)b[0] << 8) | b[1]);
  return true;
}

static bool tofWr8(uint8_t a, uint16_t reg, uint8_t v) {
  uint8_t d[3] = {(uint8_t)(reg >> 8), (uint8_t)(reg & 0xFF), v};
  return i2cWrite(a, d, 3);
}

static bool tofWr16(uint8_t a, uint16_t reg, uint16_t v) {
  uint8_t d[4] = {(uint8_t)(reg >> 8), (uint8_t)(reg & 0xFF), (uint8_t)(v >> 8), (uint8_t)(v & 0xFF)};
  return i2cWrite(a, d, 4);
}

static bool tofWr32(uint8_t a, uint16_t reg, uint32_t v) {
  uint8_t d[6] = {(uint8_t)(reg >> 8),  (uint8_t)(reg & 0xFF), (uint8_t)(v >> 24),
                  (uint8_t)(v >> 16),   (uint8_t)(v >> 8),     (uint8_t)(v & 0xFF)};
  return i2cWrite(a, d, 6);
}

static bool tofModelOk(uint8_t a) {
  uint16_t id = 0;
  return tofRd16(a, kTofRegModelId, &id) && id == kTofModelId;
}

static bool tofDataReady(uint8_t a, uint8_t intPol, bool *ready) {
  uint8_t v = 0;
  if (!tofRd8(a, kTofRegGpioTioStatus, &v)) {
    return false;
  }
  *ready = (v & 0x01) == intPol;
  return true;
}

static bool tofWaitReady(uint8_t a, uint8_t intPol, uint32_t maxMs) {
  uint32_t t0 = millis();
  while (millis() - t0 < maxMs) {
    bool ready = false;
    if (tofDataReady(a, intPol, &ready) && ready) {
      return true;
    }
    delay(1);
  }
  return false;
}

static bool tofClearInterrupt(uint8_t a) { return tofWr8(a, kTofRegIntClear, 0x01); }
static bool tofStopRanging(uint8_t a) { return tofWr8(a, kTofRegSystemStart, 0x00); }
static bool tofStartContinuous(uint8_t a) { return tofWr8(a, kTofRegSystemStart, 0x21); }

static uint16_t tofEncodeTimeout(uint32_t budgetUs4096, uint32_t macroUs, uint32_t mult) {
  uint32_t tmp = (macroUs * mult) >> 6;
  uint32_t ls = ((budgetUs4096 + (tmp >> 1)) / tmp) - 1;
  uint16_t ms = 0;
  while ((ls & 0xFFFFFF00U) != 0) {
    ls >>= 1;
    ms++;
  }
  return (uint16_t)((ms << 8) | (ls & 0xFF));
}

static bool tofSetTiming(uint8_t a, uint32_t budgetMs, uint32_t interMs) {
  uint16_t osc = 0;
  if (!tofRd16(a, kTofRegOscFreq, &osc) || osc == 0) {
    return false;
  }
  if (budgetMs < 10 || budgetMs > 200) {
    return false;
  }
  uint32_t macroUs = ((uint32_t)2304 * ((uint32_t)0x40000000 / osc)) >> 6;
  uint32_t budgetUs = budgetMs * 1000;
  if (interMs == 0) {
    if (!tofWr32(a, kTofRegInterMeasurement, 0)) {
      return false;
    }
    budgetUs -= 2500;
  } else if (interMs > budgetMs) {
    uint16_t pll = 0;
    if (!tofRd16(a, kTofRegOscCalibrate, &pll)) {
      return false;
    }
    pll &= 0x3FF;
    uint32_t im = (uint32_t)(1.055f * (float)interMs * (float)pll);
    if (!tofWr32(a, kTofRegInterMeasurement, im)) {
      return false;
    }
    budgetUs = (budgetUs - 4300) / 2;
  } else {
    return false;
  }
  uint32_t b4096 = budgetUs << 12;
  return tofWr16(a, kTofRegRangeConfigA, tofEncodeTimeout(b4096, macroUs, 16)) &&
         tofWr16(a, kTofRegRangeConfigB, tofEncodeTimeout(b4096, macroUs, 12));
}

static bool tofReadResult(uint8_t a, TofResult *r) {
  static const uint8_t kStatusMap[24] = {255, 255, 255, 5,   2,   4,   1,   7,  3,  0,  255, 255,
                                         9,   13,  255, 255, 255, 255, 10,  6,  255, 255, 11, 12};
  uint8_t b[15];
  if (!tofRd(a, kTofRegResultBlock, b, sizeof(b))) {
    return false;
  }
  uint8_t st = b[0] & 0x1F;
  r->status = st < 24 ? kStatusMap[st] : st;
  r->spad = (uint16_t)((((uint16_t)b[3] << 8) | b[4]) / 256);
  uint16_t sig = (uint16_t)((((uint16_t)b[5] << 8) | b[6]) * 8);
  uint16_t amb = (uint16_t)((((uint16_t)b[7] << 8) | b[8]) * 8);
  r->sigmaMm = (uint16_t)((((uint16_t)b[9] << 8) | b[10]) / 4);
  r->mm = (uint16_t)(((uint16_t)b[13] << 8) | b[14]);
  if (r->spad == 0) {
    r->status = 255;
    r->sigPerSpad = 0;
    r->ambPerSpad = 0;
    return true;
  }
  r->sigPerSpad = sig / r->spad;
  r->ambPerSpad = amb / r->spad;
  return true;
}

static bool tofSetOffset(uint8_t a, int16_t offsetMm) {
  return tofWr16(a, 0x001E, (uint16_t)(offsetMm * 4)) && tofWr16(a, 0x0020, 0) &&
         tofWr16(a, 0x0022, 0);
}

// Full ST SensorInit sequence with bounded waits, then our range timing and
// thresholds. Returns false on any failed transaction or timeout.
static bool tofInitDevice(uint8_t a, uint8_t *intPol, int16_t offsetMm) {
  if (!tofModelOk(a)) {
    return false;
  }
  uint32_t t0 = millis();
  for (;;) {
    uint8_t st = 0;
    if (tofRd8(a, kTofRegFirmwareStatus, &st) && st == 0x03) {
      break;
    }
    if (millis() - t0 > kTofBootMs) {
      return false;
    }
    delay(1);
  }
  for (uint8_t i = 0; i < sizeof(kTofDefaultConfig); ++i) {
    if (!tofWr8(a, (uint16_t)(0x2D + i), kTofDefaultConfig[i])) {
      return false;
    }
  }
  uint8_t gpio = 0;
  if (!tofRd8(a, kTofRegGpioHvMux, &gpio)) {
    return false;
  }
  *intPol = (gpio & 0x10) ? 0 : 1;
  if (!tofWr8(a, kTofRegSystemStart, 0x40)) {
    return false;
  }
  if (!tofWaitReady(a, *intPol, kTofVhvMs)) {
    return false;
  }
  if (!tofClearInterrupt(a) || !tofStopRanging(a)) {
    return false;
  }
  if (!tofWr8(a, kTofRegVhvLoopBound, 0x09) || !tofWr8(a, 0x000B, 0x00) ||
      !tofWr16(a, 0x0024, 0x0500)) {
    return false;
  }
  if (!tofSetTiming(a, kTofBudgetMs, 0)) {
    return false;
  }
  if (!tofWr16(a, kTofRegMinCountRate, (uint16_t)(kTofSignalKcps >> 3)) ||
      !tofWr16(a, kTofRegSigmaThresh, (uint16_t)(kTofSigmaMm << 2))) {
    return false;
  }
  return tofSetOffset(a, offsetMm);
}

// Move a VL53L4CD to a new 7-bit address (volatile until the sensor loses
// power). Verified by probing the new address.
static bool tofSetAddress(uint8_t from, uint8_t to) {
  if (!tofWr8(from, kTofRegSlaveAddr, to)) {
    return false;
  }
  delay(2);
  return i2cProbe(to);
}

// ---- ADXL345 --------------------------------------------------------------

static bool accelRd(uint8_t a, uint8_t reg, uint8_t *b, size_t n) { return i2cWriteRead(a, &reg, 1, b, n); }

static bool accelWr(uint8_t a, uint8_t reg, uint8_t v) {
  uint8_t d[2] = {reg, v};
  return i2cWrite(a, d, 2);
}

static bool accelInitDevice(uint8_t a) {
  uint8_t id = 0;
  if (!accelRd(a, 0x00, &id, 1) || id != 0xE5) {
    return false;
  }
  if (!accelWr(a, 0x2D, 0x00) ||          // POWER_CTL: standby while configuring
      !accelWr(a, 0x31, kAccelFormat) ||  // DATA_FORMAT: full resolution, +/-16 g
      !accelWr(a, 0x2C, kAccelRate) ||    // BW_RATE
      !accelWr(a, 0x2E, 0x00) ||          // INT_ENABLE: none
      !accelWr(a, 0x38, 0x80) ||          // FIFO_CTL: stream mode, 32 deep
      !accelWr(a, 0x1E, 0x00) || !accelWr(a, 0x1F, 0x00) || !accelWr(a, 0x20, 0x00) ||
      !accelWr(a, 0x2D, 0x08)) {          // POWER_CTL: measure
    return false;
  }
  uint8_t fmt = 0;
  return accelRd(a, 0x31, &fmt, 1) && fmt == kAccelFormat;
}

// ---- slot table -----------------------------------------------------------

static const char *kindName(uint8_t kind) { return kind == kKindAccel ? "accel" : "tof"; }

static void formatId(char *out, size_t n, uint8_t mux, uint8_t ch, uint8_t kind, uint8_t addr) {
  const char *suffix = (kind == kKindAccel && addr == kAccelAddrAlt) ? "1d" : "";
  if (mux == 0) {
    snprintf(out, n, "root:%s%s", kindName(kind), suffix);
  } else {
    snprintf(out, n, "mux:%x:%u:%s%s", mux, ch, kindName(kind), suffix);
  }
}

static int findSlot(uint8_t mux, uint8_t ch, uint8_t kind, uint8_t addr) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    SensorSlot &s = slots[i];
    if (s.mux != mux || s.ch != ch || s.kind != kind) {
      continue;
    }
    if (s.kind == kKindTof || s.addr == addr) {
      return i;
    }
  }
  return -1;
}

static void reportSlot(uint8_t i) {
  if (i >= slotCount) {
    return;
  }
  SensorSlot &s = slots[i];
  Serial.printf(
      "{\"v\":1,\"t\":\"sensor\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"kind\":\"%s\",\"addr\":%u,\"chip\":\"%x\",\"ok\":%s,\"err\":%u,\"init\":%u}\n",
      s.id, s.mux, s.ch, kindName(s.kind), s.addr, s.chip, s.ok ? "true" : "false", s.errors, s.reinits);
  s.reported = true;
}

static void setSlotOk(int idx, bool ok) {
  if (idx < 0) {
    return;
  }
  SensorSlot &s = slots[idx];
  bool changed = !(s.reported && s.ok == ok);
  s.ok = ok;
  if (!ok) {
    s.failures = 0;
    s.ranging = false;
    s.calActive = false;
  }
  if (changed) {
    reportSlot((uint8_t)idx);
  }
}

static int ensureSlot(uint8_t mux, uint8_t ch, uint8_t kind, uint8_t addr) {
  int idx = findSlot(mux, ch, kind, addr);
  if (idx >= 0) {
    slots[idx].seen = true;
    if (slots[idx].addr != addr) {
      slots[idx].addr = addr;
      if (slots[idx].ok) {
        setSlotOk(idx, false);
      }
    }
    return idx;
  }
  if (slotCount >= kMaxSlots) {
    return -1;
  }
  idx = slotCount++;
  SensorSlot &s = slots[idx];
  memset(&s, 0, sizeof(s));
  formatId(s.id, sizeof(s.id), mux, ch, kind, addr);
  s.mux = mux;
  s.ch = ch;
  s.kind = kind;
  s.addr = addr;
  s.seen = true;
  if (kind == kKindTof) {
    uint16_t id16 = 0;
    if (tofRd16(addr, kTofRegModelId, &id16)) {
      s.chip = id16;
    }
    s.offsetMm = (int16_t)prefs.getShort(s.id, 0);
  } else {
    uint8_t id8 = 0;
    if (accelRd(addr, 0x00, &id8, 1)) {
      s.chip = id8;
    }
  }
  return idx;
}

static void slotFailed(int idx, bool hard) {
  if (idx < 0) {
    return;
  }
  noteBusError();
  SensorSlot &s = slots[idx];
  if (s.errors < 0xFFFF) {
    s.errors++;
  }
  if (hard) {
    s.failures = kFailLimit;
  } else if (++s.failures < kFailLimit) {
    return;
  }
  setSlotOk(idx, false);
}

// ---- bus selection --------------------------------------------------------

static int countHits() {
  int n = 0;
  for (uint8_t addr = 0x70; addr <= 0x77; ++addr) {
    if (i2cProbe(addr)) {
      n++;
    }
  }
  const uint8_t roots[4] = {kTofAddr, kTofRootAlt, kAccelAddr, kAccelAddrAlt};
  for (uint8_t i = 0; i < 4; ++i) {
    if (i2cProbe(roots[i])) {
      n++;
    }
  }
  return n;
}

static bool tryPins(int sda, int scl) {
  M5.Ex_I2C.release();
  if (!i2cBegin(sda, scl)) {
    return false;
  }
  i2cSda = (uint8_t)sda;
  i2cScl = (uint8_t)scl;
  i2cPort = (sda == 13 || sda == 15 || scl == 13 || scl == 15) ? 'a' : 'b';
  return true;
}

static void pickBus(bool quiet) {
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
  if (quiet && bestN <= 0) {
    return;
  }
  Serial.printf("{\"v\":1,\"t\":\"i2c\",\"port\":\"%c\",\"sda\":%u,\"scl\":%u,\"n\":%d}\n",
                i2cPort, i2cSda, i2cScl, bestN < 0 ? 0 : bestN);
}

// ---- discovery ------------------------------------------------------------

static bool anySeen() {
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].seen) {
      return true;
    }
  }
  return false;
}


static void forgetMux(uint8_t mux) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].mux == mux) {
      slots[i].seen = false;
      setSlotOk(i, false);
    }
  }
}

// Re-probe 0x70..0x77 and reconcile the mux table. Slots behind a vanished
// mux are marked missing immediately.
static void discoverMuxes() {
  uint8_t found[8];
  uint8_t n = 0;
  for (uint8_t addr = 0x70; addr <= 0x77; ++addr) {
    if (i2cProbe(addr)) {
      found[n++] = addr;
    }
  }
  for (uint8_t i = 0; i < muxCount; ++i) {
    bool still = false;
    for (uint8_t j = 0; j < n; ++j) {
      if (found[j] == muxAddrs[i]) {
        still = true;
      }
    }
    if (!still) {
      forgetMux(muxAddrs[i]);
    }
  }
  bool changed = n != muxCount || memcmp(found, muxAddrs, n) != 0;
  memcpy(muxAddrs, found, n);
  muxCount = n;
  if (changed) {
    muxCacheValid = false;
  }
}

// Root bus: everything visible with every mux channel closed. A root
// VL53L4CD is moved to 0x2A so a 0x29 behind a mux channel never collides
// with it. Devices that cannot move fence off their address behind muxes.
static void probeRoot(bool verbose) {
  if (!muxSelect(0, 0)) {
    return;
  }
  rootTofFixed = false;
  rootAccelFixed = false;
  rootAccelAltFixed = false;
  if (i2cProbe(kTofRootAlt)) {
    ensureSlot(0, 0, kKindTof, kTofRootAlt);
  } else if (i2cProbe(kTofAddr)) {
    if (tofModelOk(kTofAddr) && tofSetAddress(kTofAddr, kTofRootAlt)) {
      ensureSlot(0, 0, kKindTof, kTofRootAlt);
      if (verbose) {
        sendLog("root VL53L4CD moved to 0x2a");
      }
    } else {
      rootTofFixed = true;
      int idx = ensureSlot(0, 0, kKindTof, kTofAddr);
      if (idx >= 0 && slots[idx].ok) {
        setSlotOk(idx, false);
      }
      if (verbose) {
        sendLog("root 0x29 device will not move; 0x29 behind muxes is not scanned");
      }
    }
  }
  if (i2cProbe(kAccelAddr)) {
    rootAccelFixed = true;
    ensureSlot(0, 0, kKindAccel, kAccelAddr);
    if (verbose && muxCount > 0) {
      sendLog("root ADXL345 at 0x53 hides 0x53 behind muxes; strap SDO high for 0x1d");
    }
  }
  if (i2cProbe(kAccelAddrAlt)) {
    rootAccelAltFixed = true;
    ensureSlot(0, 0, kKindAccel, kAccelAddrAlt);
    if (verbose && muxCount > 0) {
      sendLog("root ADXL345 at 0x1d hides 0x1d behind muxes");
    }
  }
}

static void probeChannel(uint8_t mux, uint8_t ch) {
  if (!muxSelect(mux, ch)) {
    return;
  }
  if (!rootTofFixed && i2cProbe(kTofAddr)) {
    ensureSlot(mux, ch, kKindTof, kTofAddr);
  }
  if (!rootAccelFixed && i2cProbe(kAccelAddr)) {
    ensureSlot(mux, ch, kKindAccel, kAccelAddr);
  }
  if (!rootAccelAltFixed && i2cProbe(kAccelAddrAlt)) {
    ensureSlot(mux, ch, kKindAccel, kAccelAddrAlt);
  }
}

static void markUnseenMissing() {
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (!slots[i].seen) {
      setSlotOk(i, false);
    }
  }
}

// Full synchronous discovery: pick the Grove port, enumerate muxes, then the
// root bus, then every channel. Used for host "scan" requests and link-up.
// Sensors that already deliver samples keep ranging; only new or failed
// slots go through init. Automatic empty-bus rescans stay quiet unless they
// find something.
static void discoverSensors(bool quiet) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    slots[i].seen = false;
  }
  pickBus(quiet);
  muxCount = 0;
  if (exReady) {
    discoverMuxes();
    muxCloseAll();
    probeRoot(true);
    for (uint8_t m = 0; m < muxCount; ++m) {
      for (uint8_t ch = 0; ch < 8; ++ch) {
        probeChannel(muxAddrs[m], ch);
      }
    }
    muxSelect(0, 0);
  }
  markUnseenMissing();
  for (uint8_t i = 0; i < slotCount; ++i) {
    if (slots[i].seen && !slots[i].ok) {
      slots[i].failures = 0;
    }
  }
  if (quiet && muxCount == 0 && !anySeen()) {
    return;
  }
  Serial.printf("{\"v\":1,\"t\":\"scan\",\"n\":%u}\n", slotCount);
  for (uint8_t i = 0; i < slotCount; ++i) {
    reportSlot(i);
  }
  bgCursor = 0;
}

// Incremental background probe: one position per tick so hot-plugged
// muxes and sensors appear without a host scan and without a bus hiccup.
static void backgroundProbe(uint32_t now) {
  if (!exReady || now - lastBgProbeMs < kBgProbeMs) {
    return;
  }
  lastBgProbeMs = now;
  if (bgCursor == 0) {
    discoverMuxes();
  } else if (bgCursor <= 64) {
    uint8_t m = (uint8_t)((bgCursor - 1) / 8);
    uint8_t ch = (uint8_t)((bgCursor - 1) % 8);
    if (m < muxCount) {
      probeChannel(muxAddrs[m], ch);
    }
  } else {
    probeRoot(false);
  }
  bgCursor = (uint8_t)((bgCursor + 1) % 66);
}

// ---- init / poll ----------------------------------------------------------

static bool initTofSlot(int idx) {
  SensorSlot &s = slots[idx];
  if (!muxSelect(s.mux, s.ch) || !i2cProbe(s.addr)) {
    return false;
  }
  if (!tofInitDevice(s.addr, &s.intPol, s.offsetMm)) {
    return false;
  }
  if (!tofStartContinuous(s.addr)) {
    return false;
  }
  s.ranging = true;
  s.dropFirst = true;
  s.lastDataMs = millis();
  s.nextPollMs = s.lastDataMs + kTofBudgetMs;
  return true;
}

static bool initAccelSlot(int idx) {
  SensorSlot &s = slots[idx];
  if (!muxSelect(s.mux, s.ch)) {
    return false;
  }
  if (!accelInitDevice(s.addr)) {
    return false;
  }
  s.lastDataMs = millis();
  return true;
}

// One init attempt per kInitRetryMs, round-robin over slots that were seen
// on the bus but are not yet delivering samples.
static void initPending(uint32_t now) {
  if (now - lastInitMs < kInitRetryMs || slotCount == 0) {
    return;
  }
  lastInitMs = now;
  for (uint8_t n = 0; n < slotCount; ++n) {
    uint8_t i = nextInit;
    nextInit = (uint8_t)((nextInit + 1) % slotCount);
    if (!slots[i].seen || slots[i].ok) {
      continue;
    }
    bool ok = slots[i].kind == kKindAccel ? initAccelSlot(i) : initTofSlot(i);
    slots[i].failures = 0;
    if (ok) {
      if (slots[i].everOk && slots[i].reinits < 0xFFFF) {
        slots[i].reinits++;
      }
      slots[i].everOk = true;
    }
    setSlotOk(i, ok);
    if (ok) {
      noteBusOk();
    }
    return;
  }
}

static void scanSensors(uint32_t now) {
  if (scanRequested) {
    scanRequested = false;
    discoverSensors(false);
    lastInitMs = now;
    lastEmptyRescanMs = now;
    return;
  }
  if (muxCount == 0 && !anySeen()) {
    if (now - lastEmptyRescanMs >= kEmptyRescanMs) {
      lastEmptyRescanMs = now;
      discoverSensors(true);
      lastInitMs = now;
    }
    return;
  }
  backgroundProbe(now);
  initPending(now);
}

static void calSample(SensorSlot &s, const TofResult &r, uint32_t now);

static void pollTof(uint32_t now) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    SensorSlot &s = slots[i];
    if (s.kind != kKindTof || !s.ok || !s.ranging) {
      continue;
    }
    if ((int32_t)(now - s.nextPollMs) < 0) {
      continue;
    }
    if (!muxSelect(s.mux, s.ch)) {
      slotFailed(i, false);
      continue;
    }
    bool ready = false;
    if (!tofDataReady(s.addr, s.intPol, &ready)) {
      slotFailed(i, false);
      s.nextPollMs = now + 5;
      continue;
    }
    if (!ready) {
      if (now - s.lastDataMs > kStallMs) {
        slotFailed(i, true);
        continue;
      }
      s.nextPollMs = now + 3;
      continue;
    }
    TofResult r;
    if (!tofReadResult(s.addr, &r)) {
      slotFailed(i, false);
      s.nextPollMs = now + 5;
      continue;
    }
    tofClearInterrupt(s.addr);
    s.failures = 0;
    s.lastDataMs = now;
    s.nextPollMs = now + kTofBudgetMs - 6;
    noteBusOk();
    if (s.dropFirst) {
      s.dropFirst = false;
      continue;
    }
    calSample(s, r, now);
    uint16_t mm = r.status == 0 ? r.mm : 0;
    Serial.printf(
        "{\"v\":1,\"t\":\"tof\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"mm\":%u,\"st\":%u,\"sig\":%u,\"amb\":%u,\"spad\":%u,\"sg\":%u}\n",
        s.id, s.mux, s.ch, mm, r.status, r.sigPerSpad, r.ambPerSpad, r.spad, r.sigmaMm);
  }
}

static int accelMg(int16_t raw) { return (int)raw * 39 / 10; }

// Polled every loop pass. FIFO_CTL and FIFO_STATUS come back in one read:
// FIFO_CTL proves the part still holds our configuration (a brown-out resets
// it) and FIFO_STATUS says how many 50 Hz frames wait in the 32-deep stream
// FIFO, so a slow host tick never loses a frame.
static void pollAccel(uint32_t now) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    SensorSlot &s = slots[i];
    if (s.kind != kKindAccel || !s.ok) {
      continue;
    }
    if (!muxSelect(s.mux, s.ch)) {
      slotFailed(i, false);
      continue;
    }
    uint8_t st[2];
    if (!accelRd(s.addr, 0x38, st, sizeof(st))) {
      slotFailed(i, false);
      continue;
    }
    if (st[0] != 0x80) {
      slotFailed(i, true);
      continue;
    }
    uint8_t n = st[1] & 0x3F;
    if (n == 0) {
      if (now - s.lastDataMs > kStallMs) {
        slotFailed(i, true);
      }
      continue;
    }
    if (n > 32) {
      n = 32;
    }
    for (uint8_t k = 0; k < n; ++k) {
      uint8_t b[6];
      if (!accelRd(s.addr, 0x32, b, sizeof(b))) {
        slotFailed(i, false);
        break;
      }
      s.failures = 0;
      s.lastDataMs = now;
      noteBusOk();
      int16_t rawX = (int16_t)((uint16_t)b[1] << 8 | b[0]);
      int16_t rawY = (int16_t)((uint16_t)b[3] << 8 | b[2]);
      int16_t rawZ = (int16_t)((uint16_t)b[5] << 8 | b[4]);
      Serial.printf(
          "{\"v\":1,\"t\":\"accel\",\"id\":\"%s\",\"mux\":%u,\"ch\":%u,\"x\":%d,\"y\":%d,\"z\":%d}\n",
          s.id, s.mux, s.ch, accelMg(rawX), accelMg(rawY), accelMg(rawZ));
    }
  }
}

// ---- offset calibration ---------------------------------------------------

static void reportCal(SensorSlot &s, bool ok, int avg, uint8_t n) {
  Serial.printf("{\"v\":1,\"t\":\"cal\",\"id\":\"%s\",\"ok\":%s,\"offset\":%d,\"avg\":%d,\"n\":%u}\n",
                s.id, ok ? "true" : "false", (int)s.offsetMm, avg, n);
}

// Stop ranging, write the offset, restart. The next result is discarded.
static bool tofApplyOffset(SensorSlot &s) {
  if (!muxSelect(s.mux, s.ch) || !tofStopRanging(s.addr) || !tofSetOffset(s.addr, s.offsetMm) ||
      !tofStartContinuous(s.addr)) {
    return false;
  }
  s.dropFirst = true;
  s.lastDataMs = millis();
  s.nextPollMs = s.lastDataMs + kTofBudgetMs;
  return true;
}

// Host request: {"t":"cal","id":"mux:70:0:tof","mm":100}. mm 0 clears the
// stored offset. Otherwise the offset is zeroed, kCalSamples valid ranges are
// averaged by pollTof, and offset = target - average is applied and stored.
static void startCalibration(const String &id, int targetMm) {
  for (uint8_t i = 0; i < slotCount; ++i) {
    SensorSlot &s = slots[i];
    if (s.kind != kKindTof || strcmp(s.id, id.c_str()) != 0) {
      continue;
    }
    if (targetMm <= 0) {
      s.calActive = false;
      s.offsetMm = 0;
      prefs.remove(s.id);
      bool ok = s.ok && tofApplyOffset(s);
      reportCal(s, ok, 0, 0);
      return;
    }
    if (!s.ok || !s.ranging) {
      reportCal(s, false, 0, 0);
      return;
    }
    int16_t keep = s.offsetMm;
    s.offsetMm = 0;
    if (!tofApplyOffset(s)) {
      s.offsetMm = keep;
      slotFailed(i, true);
      reportCal(s, false, 0, 0);
      return;
    }
    s.calActive = true;
    s.calCount = 0;
    s.calSum = 0;
    s.calTarget = (uint16_t)targetMm;
    s.calStartMs = millis();
    return;
  }
  Serial.printf("{\"v\":1,\"t\":\"cal\",\"id\":\"%s\",\"ok\":false,\"offset\":0,\"avg\":0,\"n\":0}\n", id.c_str());
}

static void calSample(SensorSlot &s, const TofResult &r, uint32_t now) {
  if (!s.calActive) {
    return;
  }
  if (r.status == 0) {
    s.calSum += r.mm;
    s.calCount++;
  }
  if (s.calCount >= kCalSamples) {
    int avg = (int)(s.calSum / s.calCount);
    s.calActive = false;
    s.offsetMm = (int16_t)((int)s.calTarget - avg);
    bool ok = tofApplyOffset(s);
    if (ok) {
      prefs.putShort(s.id, s.offsetMm);
    }
    reportCal(s, ok, avg, s.calCount);
    return;
  }
  if (now - s.calStartMs > kCalTimeoutMs) {
    s.calActive = false;
    int avg = s.calCount ? (int)(s.calSum / s.calCount) : 0;
    s.offsetMm = (int16_t)prefs.getShort(s.id, 0);
    tofApplyOffset(s);
    reportCal(s, false, avg, s.calCount);
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
  if (jsonHasType(s, "cal")) {
    startCalibration(jsonString(s, "id"), jsonInt(s, "mm", 0));
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
  Serial.setTxTimeoutMs(5);
  prefs.begin("tofcal", false);
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

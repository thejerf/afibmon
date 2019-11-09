/*

  This file contains code for dealing with packets of heart information
  obtained from the EKG. Cross-check with afibmon/heartmon/records.go.

  I'm OK with C the language, but I've never learned much about the
  compilers and such, and I've got no interest in fighting with the C
  compiler to do this "properly" as a separately-compiled module, so this
  is really a "header-only library", just one I feel ought to be given a .c
  extension. What's the point in fighting with the compiler when this just
  works?

  I am making a core assumption in this process, which is that the Wifi
  chip, being intended for an embedded use (I hope...) is doing some
  management of TCP on its own. I am running on the theory that I can
  generally send a (small) packet somewhere and we are not synchronously
  waiting for it to be delivered, which would render the whole Wifi thing
  generally useless for a lot of things, so I hope it's a safe assumption.

*/

const char TIMESTAMP = 1;
const char HEARTDATA = 2;
const char ERROR = 3;
const int INITIAL = -1;
const int ERROR_PKT = -2;
const int BUFFER_SIZE = 16384;

void everyTwoSeconds();
void appendError(char*);

// The next packet we are constructing
char nextPacket[BUFFER_SIZE];

// Where we are in this packet. -1 indicates that we're at the
// beginning. -2 indicates an error state because we overflowed the buffer,
// and we should wait for the next initial packet and reset.
int nextPacketIdx = -1;

// At the sampling rate this gives us, this should be rather embarassingly
// more than enough.
unsigned short heartdata[1000];

// The last index of the valid heartdata. -1 indicates no valid data.
int heartdataIdx = -1;

// calling this will configure the RTC to begin the alarm chain.
void startSendingPackets () {
  rtc.attachInterrupt(everyTwoSeconds);
  rtc.enableAlarm(rtc.MATCH_SS);
  rtc.setAlarmSeconds(rtc.getSeconds() + 5);
}

void appendDatum(unsigned short datum) {
  heartdataIdx++;
  if (heartdataIdx > 1000) {
    appendError("overflow heart data");
    return;
  }
  heartdata[heartdataIdx] = datum;
}

void newPacket() {
  if (nextPacketIdx != ERROR_PKT) {
    return;
  }

  nextPacketIdx = INITIAL;
  appendError("PACKET OVERFLOW");
}

void pktWrite(char c) {
  if (nextPacketIdx == ERROR_PKT) {
    return;
  }

  nextPacketIdx++;
  if (nextPacketIdx > BUFFER_SIZE) {
    // Somehow, we have actually overflowed this entire packet. What the
    // heck. Well, panic out the serial port and try just throwing away the
    // entire packet to date, because what else is there to do?
    // Well, I guess we can try to literally squeal.
    tone(piezoPin, NOTE_G5, 5000);
    nextPacketIdx = ERROR_PKT;
  }

  nextPacket[nextPacketIdx] = c;
}

void appendError(char* s) {
  newPacket();
  pktWrite(ERROR);
  int l = strlen(s);
  pktWrite(l / 256);
  pktWrite(l % 256);

  for (int i = 0; i < l; i++) {
    pktWrite(s[i]);
  }
}

void appendTimestamp() {
  newPacket();
  // The documentation claims this is "periodically" fetched from the NTP
  // server, which tends to imply to me this is not *synchronously*
  // fetching the time every time I call this method. I'm not sure how far
  // I trust documentation that can't even document the type of the
  // returned value, though. :-/
  unsigned long epoch = WiFi.getTime();

  pktWrite(TIMESTAMP);
  pktWrite(0);
  pktWrite(4);
  pktWrite(epoch >> 24);
  pktWrite((epoch >> 16) & 0xff);
  pktWrite((epoch >> 8) & 0xff);
  pktWrite(epoch & 0xff);
}

void everyTwoSeconds () {
  // We need to append the timestamp, append the heart data packet, send
  // it, and reset the index, as long as there is any heart data.
  if (heartdataIdx > 0) {
    appendTimestamp();

    pktWrite(HEARTDATA);
    unsigned short length = ((unsigned short)(heartdataIdx) + 1) * 2;
    pktWrite(length / 256);
    pktWrite(length & 0xff);
    for (int i = 0; i <= heartdataIdx; i++) {
      pktWrite(heartdata[i] / 256);
      pktWrite(heartdata[i] & 0xff);
    }

    // I hope it's OK to write a Wifi packet in an interrupt handler.
    // it should just be pushing on to a buffer if this is going to work at
    // all...
    client.write(nextPacket, nextPacketIdx + 1);
    nextPacketIdx = INITIAL;
    heartdataIdx = INITIAL;
    Serial.println("Packet sent");
  }

  // reset the alarm so it fires again regardless of what happens
  rtc.setAlarmSeconds((rtc.getSeconds() + 2) % 60);
}

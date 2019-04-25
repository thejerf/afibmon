/*

  heart_monitor.ino - part of https://github.com/thejerf/afibmon

  This code is cobbled together from several public-domain example code
  bits, for the RTC code, the code to hook to Wifi, the NTP lookup for
  time, etc.

*/

/*

 * TODO: If connection is lost, attempt to reconnect, preferably
   without losing heart data as long as it reconnects quickly enough.

*/

#include <SPI.h>
#include <WiFiNINA.h> //Include this instead of WiFi101.h as needed
#include <WiFiUdp.h>
#include <RTCZero.h>

// If you error out on this line, look at the README for this repo
// again. You have to create a file to contain your WIFI secrets.
// Don't accidentally commit it to GitHub! .gitignore is configured in this
// directory but you can still override it.
#include "secrets.h"

#include "pitches.h"
#include "packets.c"

RTCZero rtc;

WiFiClient client;

int status = WL_IDLE_STATUS;

const int TZ_OFFSET = 2; //change this to adapt it to your time zone

int connected = 0;

int piezoPin = 0;

void setup() {
  Serial.begin(115200);
  Serial.print("Hello!");
  tone(piezoPin, NOTE_C3, 500);
  Serial.print("generated the tone");

  if (WiFi.status() == WL_NO_SHIELD) {
    while (1) {
      blink(200, 100);
      tone(piezoPin, NOTE_C3, 250);
      blink(500, 100);
    }
  }

  // attempt to connect to WiFi network:
  int attempts = 0;
  while ( status != WL_CONNECTED && attempts < 5) {
    Serial.print("Attempting to connect to SSID: ");
    Serial.println(ssid);
    // Connect to WPA/WPA2 network. Change this line if using open or WEP network:
    status = WiFi.begin(ssid, pass);

    // wait 10 seconds for connection:
    delay(10000);
    attempts++;
  }

  tone(piezoPin, NOTE_C3, 250);
  delay(250);
  tone(piezoPin, NOTE_E3, 250);
  delay(250);

  if (status != WL_CONNECTED) {
    while (1) {
      blink(200, 100);
      tone(piezoPin, NOTE_C3, 500);
      blink(1000, 100);
    }
  }

  blinkr(50, 50, 2);
  tone(piezoPin, NOTE_C3, 250);
  delay(250);
  tone(piezoPin, NOTE_E3, 250);
  delay(250);
  tone(piezoPin, NOTE_F3, 250);
  delay(250);

  // print out the connection status for debugging purposes
  printWiFiStatus();

  rtc.begin();

  unsigned long epoch;
  int numberOfTries = 0, maxTries = 6;
  do {
    epoch = WiFi.getTime();
    numberOfTries++;
  }
  while ((epoch == 0) && (numberOfTries < maxTries));

  if (numberOfTries > maxTries) {
    Serial.print("NTP unreachable!!");
    while (1) {
      blinkr(200, 100, 2);
      blink(1000, 100);
    }
  }
  else {
    Serial.print("Epoch received: ");
    Serial.println(epoch);
    rtc.setEpoch(epoch);

    Serial.println();
  }

  blinkr(100, 50, 2);

  tone(piezoPin, NOTE_C3, 250);
  delay(250);
  tone(piezoPin, NOTE_E3, 250);
  delay(250);
  tone(piezoPin, NOTE_F3, 250);
  delay(250);
  tone(piezoPin, NOTE_G3, 250);
  delay(250);

  connected = client.connect(monitorServer, monitorPort);
  if (!connected) {
    Serial.print("Could not connect to monitor server\n");
    while (1) {
      tone(piezoPin, NOTE_F3, 250);
      blinkr(200, 100, 3);
      blink(1000, 100);
    }
  }

  tone(piezoPin, NOTE_C3, 250);
  delay(250);
  tone(piezoPin, NOTE_E3, 250);
  delay(250);
  tone(piezoPin, NOTE_F3, 250);
  delay(250);
  tone(piezoPin, NOTE_G3, 250);
  delay(250);
  tone(piezoPin, NOTE_C4, 250);
  delay(250);

  // Set up the every-five-second interrupt handler to send data.
  startSendingPackets();
}

// Blink the onboard LED synchronously for the given on time and off time,
// in milliseconds.
void blink(int onMS, int offMS) {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(onMS);
  digitalWrite(LED_BUILTIN, LOW);
  delay(offMS);
}

// Blink the onboard LED synchronously for the given on and off time, for
// the given number of repeats.
void blinkr(int onMS, int offMS, int repeat) {
  for (int i = 0; i < repeat; i++) {
    blink(onMS, offMS);
  }
}

uint16 value = 0;

void loop() {
  appendDatum(value++);
  delay(25);
}

void printTime()
{
  print2digits(rtc.getHours() + TZ_OFFSET);
  Serial.print(":");
  print2digits(rtc.getMinutes());
  Serial.print(":");
  print2digits(rtc.getSeconds());
  Serial.println();
}

void printDate()
{
  Serial.print(rtc.getDay());
  Serial.print("/");
  Serial.print(rtc.getMonth());
  Serial.print("/");
  Serial.print(rtc.getYear());

  Serial.print(" ");
}


void printWiFiStatus() {
  // print the SSID of the network you're attached to:
  Serial.print("SSID --: ");
  Serial.println(WiFi.SSID());

  // print your WiFi shield's IP address:
  IPAddress ip = WiFi.localIP();
  Serial.print("IP Address --: ");
  Serial.println(ip);

  // print the received signal strength:
  long rssi = WiFi.RSSI();
  Serial.print("signal strength (RSSI):");
  Serial.print(rssi);
  Serial.println(" dBm");
}

void print2digits(int number) {
  if (number < 10) {
    Serial.print("0");
  }
  Serial.print(number);
}

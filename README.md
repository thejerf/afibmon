# afibmon

A repository holding the code I'm writing for my atrial fibrillation monitor.

At the moment this is incomplete. When it is completed more information will be placed here.

# Note

Really, all open source users should understand that the source of their open source code has no obligation to help them out.

However, this is extra super ultra true of this repo. If you have trouble with the "payload" code that takes the heart data and processes I'm interested in hearing from you. If you have contributions to make for additional outputs, I'm interested in hearing from you. If you know better ways to process the data than my somewhat hack and slash approach, I'm interested in hearing from you. If you have sufficient information to figure out how to monitor for other conditions, I'm interested in hearing from you but bear in mind you're going to be doing a lot of the work.

However:

1. I am NOT YOUR DOCTOR. If you are wondering if you have atrial fibrillation or other heart issues, you are in the WRONG PLACE. Go to your doctor. This is the only answer you will get for anything that involves medical judgment.
1. I am not available to help you with issues relating to getting your Arduino or Raspberry Pi going. (I'm not an expert anyhow; what I know about them is pretty much only and exactly what was necessary to get this project going.)
1. I am STILL NOT YOUR DOCTOR.
1. If for any reason you are asking a question that you believe needs to be answered immediately, I AM STILL NOT YOUR DOCTOR. Go to the emergency room.
1. The license used by this code already disclaims obligations. But let me reiterate it again here; this is ALL AT YOUR OWN RISK.

Feel free to play with this for a bit, but if you're depending on it for anything that's on you.

# Hardware Parts List

1. An Arduino MKR1010 board.
1. An AD8232 Arduino shield.
1. A passive buzzer.
1. Eventually, some form of case to put this in.

# Software Install

I'm not sure how to package this better. Suggestions welcome.

1. Install the Arduino IDE. Getting it from Linux distro packaging may be
   too old.
2. In the Arduino IDE, install any necessary package to speak to your
   board, and install WifiNINA and RTCZero. (Despite the warning that it
   only works on the MKR1000 it unsurprisingly seems to work on the MKR1010
   as well.
3. In afibmonitor/heart_monitor, create a new file with the following:

       char ssid[] = "YOURSSID";
       char pass[] = "YOURPASSWORD"
       char monitorServer[] = "YOUR_MONITOR_SERVER";
       int monitorPort = 18999; // your monitor port; this is default

   This will configure the sketch to hook to your Wifi. This is set up
   assuming WPA2; you're on your own if you're still on WEP, as I can't
   test it. We'll get to the monitor server in a moment.


# My Story

I'll expand this later, but the short version is, I'm 38 and my heart does things it shouldn't. I have been to a Real Doctor and they have declared that what's happening to me is "perfectly normal cardiac events". So let's call it "subclinical". Subclinical my issues may be, but they are still seriously affecting my quality of life. I believe (but can't prove) I have paroxysmal fibrillation. For some people, such as my mother, fibrillation turns on like a switch and they stay in it continuously. I do not. In particular, it only affects me while I'm asleep or nearly asleep. Consequently it was a long time before I even realized I had a problem, until it got bad enough for me to notice it while trying to sleep, but before falling asleep.

In addition, I have Celiac disease. "Everyone knows" that means I can't have wheat, but more subtle than that and in many ways worse than that, a Celiac sufferer can have a hard time absorbing nutrients. Even if I nominally eat enough of some nutrients I can still have deficiencies.

It is reasonably well established by science that certain substances can help fix fibrillation. So there are even treatments I can do in the middle of the night that can suppress fibrillation as it is occurring. To a first approximation, sleep I get while fibrillating does me no good, so having two hours of fibrillation in the middle of the night is like missing two hours sleep. To fix this, I need a monitor that will tell me relatively quickly when I'm fibrillating, but there does not seem to be any such device on the market. You can buy heart monitors, but either because the market can't sustain anything terribly useful or for regulatory reasons, you can not get anything that works this way. Most heart devices are optimized for the case of "Something bad is happening to me _right now_, what is it?" They take 30 second samples and send them directly to a cardiologist for identification. That's great. If you've been reading along up to this point and that sounds useful to you, I recommend checking around on Amazon for them. All things considered, they're not that expensive to just try out; depending on your insurance it could even be cheaper than the copay on one office visit for you.

But they don't do me any good. I need continual monitoring, and I need it _while I'm asleep_. Pushing buttons to get a reading is a non-starter!

You can also get longer term monitors, but they don't have any mechanism that I can see for accessing the data and processing it immediately. Consequently those are far less useful, if indeed they are useful at all. (Given how many hundreds of dollars they cost I haven't tried them out.) I _could_ improve over time with something that could retrospectively tell me how my night went, but there's no option to have sub-minute notifications at all, which is even better.

As I write this, I'm still prototyping and fiddling. No idea if this is going to work or not. But we'll see.


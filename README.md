![intro](/pics/IMG_2380.gif)

# Info
In short this is just a notification device. It's client listens to MacOS Notification Daemon events on your Mac and transmits all received notifications to the device, then it will scroll text trough the led diodes on the device front.

# Background info
Led matrix uses modified library from https://github.com/adrianh-za/go-max7219-rpi listens on specified port for http api commands. You can write any compatible client for any OS to use this device displaying your notifications in scrolling text way. It can also display time as configured in the **config.json**. I've also already made MacOS simple client that reads Notification Center SQLite database and shows new notifications depending on the MAX(id). Also, i've made a Lithuanian font support it can display text in UTF-8 unicode characters, as shown in the picture below:

![ltfonts](/pics/IMG_2830.jpeg)

The daemon is smart enough and has queue manager, if you send two text messages at once it will be displayed in sequence, not disturbing each other. It also have built-in HOTSPOT wifi, there are two GPIO pins connected with cables if you fastly connect them together for 8 times the hotsport wifi will start, it was designed to be used as emergency with button attached. When the devices are moved from one environment to another it can be easily setup because after wifi hotspot is started the web server is also started at port 80. You can then insert another wifi ap details such as wifi SSID and password to connect to.

This project source is completely written in GoLang and can be used in another projects such as big LED light boards for the advertising products and so on, with small additional work new fonts and Languages can be implemented easily. Daemon also uses avahi technology to be easily discovered on the network.
![avahi](/pics/zeroconf.png)

The MacClient also has abillity to detect device on the network on the fly, no need to specify ip address and/or connection details manually, but i can be specified to speed things up..

# What i learned from this project?
It's been fun to learn how to programmatically setup WIFI hotspot, [AVAHI](https://www.avahi.org) zero configuration networking and improve programming skills on [GoLang](https://golang.org).


# Hardware requirements
* [OrangePI Lite](https://www.aliexpress.com/item/1005002557347741.html) (RaspberryPI or clones should work fine).
* [DC Power adaptor](https://www.aliexpress.com/item/32961533195.html) for OrangePI.
* One or more Max7219 [8x8 Led matrix](https://www.aliexpress.com/item/32580532205.html).
* Some M-F [gpio wires](https://www.aliexpress.com/item/32921454163.html).
* No soldering are required at all.

# STL for 3D printing

![stl](/pics/IMG_2344.jpeg)

* [iMac holder](/stl/imac_holder.stl)
* [LED Holder box](/stl/led_holder_box.stl)


# Hardware setup

```
Board Pin	Name	Remarks		RPi Pin		RPi Function
1	        VCC	+5V Power	2		5V0
2	        GND	Ground		6		GND
3	        DIN	Data In		19		GPIO 10 (MOSI)
4	        CS	Chip Select	24		GPIO 8 (SPI CE0)
5	        CLK	Clock		23		GPIO 11 (SPI CLK)
```

# OrangePI Board OS setup
Open /boot/armbianEnv.txt and add these lines:
```
overlay_prefix=sun8i-h3
overlays=spi-spidev
param_spidev_spi_bus=0
```
# Code setup on OrangePI device
```
git clone https://github.com/e1z0/go-leds
cd go-leds
make deps && make
```
Copy systemd services led.service and Setup/led_setup.service to /etc/systemd/system (change original path to downloaded code) and start/enable them
```
systemctl enable led_setup && systemctl enable led
systemctl start led_setup && systemctl start led
```



## Compile MacOS Client
```
cd MacClient
make deps
make
make distribution
```

Launch ledofication.app the log file is at ledofication.app/Contents/MacOS/ledofication.log


# More pictures

[pic1](/pics/IMG_2337.jpeg)

[pic2](/pics/IMG_2372.jpeg)

[pic3](/pics/IMG_2373.jpeg)

[pic4](/pics/IMG_2376.jpeg)

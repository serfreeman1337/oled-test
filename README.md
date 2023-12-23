# OLED Test
A test project to add an OLED display showing network stats and room temperature to a single-board computer.

![image](https://github.com/serfreeman1337/oled-test/assets/2133936/87ae96e4-98e9-4373-a265-dce07bf42a0c)

Components:
- [CH347 development board](https://github.com/wuxx/USB-HS-Bridge)
- SSD1306 SPI OLED Display
- SHT4x Humidity and Temperature Sensor
- TSL2591 Light Sensor

At night, the lux sensor readings are used to turn off the display when there is no light, so it won't distract you. 

# hardware-exporter
hardware metrics exporter

you can use this exporter to get server's hardware info.

# how to use this exporter
1. modify the file of config.ini
   You should fill in the server hardware information
2. Compile the program
   go build hardware_exporter
3. make hardware-exporter image
   We have written the dockerfile file for you and just exec the fllowing command to generate image.
   ```shell
   docker build .

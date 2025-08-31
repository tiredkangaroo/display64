from PIL import Image
from io import BytesIO
import os, sys, socket
import struct

debug = True if os.environ.get('DEBUG') == 'true' else False

matrix = None
if not debug:
    sys.path.append('/home/pi/rpi-rgb-led-matrix/bindings/python')
    from rgbmatrix import RGBMatrix, RGBMatrixOptions, graphics
    options = RGBMatrixOptions()
    options.brightness = 50
    options.rows = 64
    options.cols = 64
    options.chain_length = 1
    options.parallel = 1
    options.gpio_slowdown = 1
    options.pwm_dither_bits  = 1
    options.pwm_lsb_nanoseconds = 90
    options.hardware_mapping = 'adafruit-hat-pwm'
    matrix = RGBMatrix(options = options)

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
    s.bind(('0.0.0.0', 14366))
    s.listen()
    print("Listening for connections...")
    while True:
        conn, addr = s.accept()
        with conn:
            print(f"Connected by {addr}")
            while True:
                length = conn.recv(8)
                if not length:
                    break
                length = struct.unpack('>Q', length)[0]
                print(f"Image data length: {length}")

                data = conn.recv(length) # receive image data
                if not data:
                    break
                img = Image.open(BytesIO(data)) # make pil image (no conversions needed to RGB, go will handle that)
                if debug:
                    img.show()
                else:
                    matrix.SetImage(img, unsafe=False)
            print(f"Connection with {addr} closed")
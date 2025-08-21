from flask import Flask, request, jsonify
from PIL import Image
from io import BytesIO
import os, sys

debug = True if os.environ.get('DEBUG') == 'true' else False

matrix = None
if not debug:
    sys.path.append('/home/pi/rpi-rgb-led-matrix/bindings/python')
    from rgbmatrix import RGBMatrix, RGBMatrixOptions, graphics
    options = RGBMatrixOptions()
    options.brightness = 80
    options.rows = 64
    options.cols = 64
    options.chain_length = 1
    options.parallel = 1
    options.gpio_slowdown = 3
    options.pwm_dither_bits  = 1
    options.pwm_lsb_nanoseconds = 90
    options.hardware_mapping = 'adafruit-hat-pwm'
    matrix = RGBMatrix(options = options)

app = Flask(__name__)

@app.route('/use', methods=['POST'])
def upload_file():
    if 'file' not in request.files:
        return jsonify({'error': 'No file part'}), 422

    file = request.files['file']
    if file.filename == '':
        return jsonify({'error': 'No selected file'}), 422

    try:
        img = Image.open(file)
        if debug:
            img.show()
        else:
            img.thumbnail((matrix.width, matrix.height), Image.LANCZOS)
            print(f"Image size after resize: {img.size}")
            print(f"{matrix}")
            matrix.SetImage(img.convert("RGB"), unsafe=False)
        return jsonify({'message': 'Image successfully processed'}), 200
    except Exception as e:
        print("invalid image file", e)
        return jsonify({'error': 'Invalid image file'}), 422

if __name__ == '__main__':
    app.run(debug=debug, port=14366, host='0.0.0.0', use_reloader=False)
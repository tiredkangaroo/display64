from flask import Flask, request, jsonify
from PIL import Image
from io import BytesIO
import os

debug = True if os.environ.get('DEBUG') == 'true' else False

matrix = None
if not debug:
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
        img = Image.open(BytesIO(file.read()))
        if debug:
            img.show()
        else:
            image.thumbnail((matrix.width, matrix.height), Image.ANTIALIAS)
            matrix.SetImage(image)
        return jsonify({'message': 'Image successfully processed'}), 200
    except Exception as e:
        return jsonify({'error': 'Invalid image file'}), 422

if __name__ == '__main__':
    app.run(debug=True, port=14366, host='0.0.0.0')

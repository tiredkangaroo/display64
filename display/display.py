from flask import Flask, request, jsonify
from PIL import Image
from io import BytesIO

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
        img.show()
        return jsonify({'message': 'Image successfully processed'}), 200
    except Exception as e:
        return jsonify({'error': 'Invalid image file'}), 422

if __name__ == '__main__':
    app.run(debug=True, port=14366)

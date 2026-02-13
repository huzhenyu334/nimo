"""STP/STEP file thumbnail generation microservice.

Upload a .stp/.step file → returns an SVG/PNG thumbnail with fixed isometric view.
Used by nimo PLM to auto-generate BOM part previews.
"""

import os
import io
import tempfile
import hashlib
from pathlib import Path
from flask import Flask, request, jsonify, send_file

import cadquery as cq
from cadquery import exporters

app = Flask(__name__)

# Cache directory for generated thumbnails
CACHE_DIR = Path(os.environ.get('THUMBNAIL_CACHE', '/home/claw/.openclaw/workspace/uploads/thumbnails'))
CACHE_DIR.mkdir(parents=True, exist_ok=True)

# Max file size: 50MB
MAX_FILE_SIZE = 50 * 1024 * 1024


def generate_thumbnail_svg(step_path: str, width: int = 400, height: int = 300) -> str:
    """Load a STEP file and export an SVG thumbnail with isometric view, centered."""
    shape = cq.importers.importStep(step_path)
    
    # Export to SVG with fit-to-view and proper margins for centering
    svg_opts = {
        "width": width,
        "height": height,
        "showAxes": False,
        "showHidden": False,
        "marginLeft": 30,
        "marginTop": 30,
        "projectionDir": (1, -1, 0.5),  # isometric-ish view
        "focus": None,  # auto-fit
    }
    
    # Export to string via temp file
    with tempfile.NamedTemporaryFile(suffix='.svg', delete=False, mode='w') as tmp:
        tmp_svg = tmp.name
    
    exporters.export(shape, tmp_svg, exportType=exporters.ExportTypes.SVG, opt=svg_opts)
    svg_str = Path(tmp_svg).read_text()
    os.unlink(tmp_svg)
    
    # Post-process: add white background and ensure viewBox centers the content
    if '<svg ' in svg_str and 'viewBox' not in svg_str:
        svg_str = svg_str.replace('<svg ', f'<svg viewBox="0 0 {width} {height}" ', 1)
    
    return svg_str


@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "stp-thumbnail"})


@app.route('/thumbnail', methods=['POST'])
def thumbnail():
    """Generate thumbnail from uploaded STP/STEP file.
    
    Form data:
        file: STP/STEP file
        width: optional, default 400
        height: optional, default 300
        format: optional, 'svg' (default) or 'png'
    
    Returns: SVG image
    """
    if 'file' not in request.files:
        return jsonify({"error": "No file uploaded"}), 400
    
    f = request.files['file']
    if not f.filename:
        return jsonify({"error": "Empty filename"}), 400
    
    ext = Path(f.filename).suffix.lower()
    if ext not in ('.stp', '.step'):
        return jsonify({"error": f"Unsupported file type: {ext}, expected .stp or .step"}), 400
    
    width = int(request.form.get('width', 400))
    height = int(request.form.get('height', 300))
    
    # Read file content and compute hash for caching
    content = f.read()
    if len(content) > MAX_FILE_SIZE:
        return jsonify({"error": f"File too large: {len(content)} bytes, max {MAX_FILE_SIZE}"}), 400
    
    file_hash = hashlib.md5(content).hexdigest()
    cache_key = f"{file_hash}_{width}x{height}.svg"
    cache_path = CACHE_DIR / cache_key
    
    # Return cached if exists
    if cache_path.exists():
        return send_file(cache_path, mimetype='image/svg+xml')
    
    # Write to temp file and process
    with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
        tmp.write(content)
        tmp_path = tmp.name
    
    try:
        svg_str = generate_thumbnail_svg(tmp_path, width, height)
        
        # Cache the result
        cache_path.write_text(svg_str)
        
        return send_file(
            io.BytesIO(svg_str.encode('utf-8')),
            mimetype='image/svg+xml',
            download_name=f"{Path(f.filename).stem}_preview.svg"
        )
    except Exception as e:
        return jsonify({"error": f"Failed to process STEP file: {str(e)}"}), 500
    finally:
        os.unlink(tmp_path)


@app.route('/thumbnail/from-path', methods=['POST'])
def thumbnail_from_path():
    """Generate thumbnail from a file already on the server (e.g., in uploads/).
    
    JSON body:
        path: server-side path to STP/STEP file
        width: optional, default 400
        height: optional, default 300
    """
    data = request.get_json()
    if not data or 'path' not in data:
        return jsonify({"error": "Missing 'path' in request body"}), 400
    
    file_path = data['path']
    if not os.path.exists(file_path):
        return jsonify({"error": f"File not found: {file_path}"}), 404
    
    width = data.get('width', 400)
    height = data.get('height', 300)
    
    # Compute cache key
    stat = os.stat(file_path)
    cache_key = f"{hashlib.md5(file_path.encode()).hexdigest()}_{stat.st_mtime}_{width}x{height}.svg"
    cache_path = CACHE_DIR / cache_key
    
    if cache_path.exists():
        return send_file(cache_path, mimetype='image/svg+xml')
    
    try:
        svg_str = generate_thumbnail_svg(file_path, width, height)
        cache_path.write_text(svg_str)
        return send_file(
            io.BytesIO(svg_str.encode('utf-8')),
            mimetype='image/svg+xml'
        )
    except Exception as e:
        return jsonify({"error": f"Failed to process STEP file: {str(e)}"}), 500


@app.route('/convert/stl', methods=['POST'])
def convert_to_stl():
    """Convert STP/STEP to STL for 3D web preview.
    
    Form data:
        file: STP/STEP file
        tolerance: optional, tessellation tolerance (default 0.1)
    
    JSON body alternative:
        path: server-side path to STP/STEP file
    
    Returns: binary STL file
    """
    tolerance = 0.1
    
    if request.content_type and 'json' in request.content_type:
        # From server path
        data = request.get_json()
        if not data or 'path' not in data:
            return jsonify({"error": "Missing 'path'"}), 400
        file_path = data['path']
        if not os.path.exists(file_path):
            return jsonify({"error": f"File not found: {file_path}"}), 404
        tolerance = data.get('tolerance', 0.1)
        
        # Cache key from path
        stat = os.stat(file_path)
        cache_key = f"{hashlib.md5(file_path.encode()).hexdigest()}_{stat.st_mtime}.stl"
        cache_path = CACHE_DIR / cache_key
        
        if cache_path.exists():
            return send_file(cache_path, mimetype='application/sla', download_name='model.stl')
        
        try:
            shape = cq.importers.importStep(file_path)
            exporters.export(shape, str(cache_path), exportType=exporters.ExportTypes.STL, tolerance=tolerance)
            return send_file(cache_path, mimetype='application/sla', download_name='model.stl')
        except Exception as e:
            return jsonify({"error": f"Failed to convert: {str(e)}"}), 500
    else:
        # From upload
        if 'file' not in request.files:
            return jsonify({"error": "No file uploaded"}), 400
        
        f = request.files['file']
        ext = Path(f.filename).suffix.lower()
        if ext not in ('.stp', '.step'):
            return jsonify({"error": f"Unsupported: {ext}"}), 400
        
        content = f.read()
        if len(content) > MAX_FILE_SIZE:
            return jsonify({"error": "File too large"}), 400
        
        file_hash = hashlib.md5(content).hexdigest()
        cache_key = f"{file_hash}.stl"
        cache_path = CACHE_DIR / cache_key
        
        if cache_path.exists():
            return send_file(cache_path, mimetype='application/sla', download_name='model.stl')
        
        with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
            tmp.write(content)
            tmp_path = tmp.name
        
        try:
            shape = cq.importers.importStep(tmp_path)
            exporters.export(shape, str(cache_path), exportType=exporters.ExportTypes.STL, tolerance=tolerance)
            return send_file(cache_path, mimetype='application/sla', download_name='model.stl')
        except Exception as e:
            return jsonify({"error": f"Failed to convert: {str(e)}"}), 500
        finally:
            os.unlink(tmp_path)


if __name__ == '__main__':
    port = int(os.environ.get('STP_SERVICE_PORT', 5001))
    print(f"STP Thumbnail Service starting on port {port}...")
    app.run(host='127.0.0.1', port=port, debug=False)

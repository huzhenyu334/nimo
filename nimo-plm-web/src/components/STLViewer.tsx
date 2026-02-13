import React, { useRef, useEffect, useState } from 'react';
import { Modal, Spin, Alert } from 'antd';
import * as THREE from 'three';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export interface STLViewerProps {
  open: boolean;
  onClose: () => void;
  fileUrl: string;
  fileName: string;
}

const STLViewer: React.FC<STLViewerProps> = ({ open, onClose, fileUrl, fileName }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    if (!open || !containerRef.current) return;

    setLoading(true);
    setError(null);

    const container = containerRef.current;
    const width = container.clientWidth || 780;
    const height = 500;

    // Scene
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf0f0f0);

    // Camera
    const camera = new THREE.PerspectiveCamera(50, width / height, 0.1, 10000);

    // Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(width, height);
    renderer.setPixelRatio(window.devicePixelRatio);
    container.appendChild(renderer.domElement);

    // Lights
    const ambientLight = new THREE.AmbientLight(0x666666);
    scene.add(ambientLight);

    const dirLight1 = new THREE.DirectionalLight(0xffffff, 1.0);
    dirLight1.position.set(1, 1, 1).normalize();
    scene.add(dirLight1);

    const dirLight2 = new THREE.DirectionalLight(0xffffff, 0.5);
    dirLight2.position.set(-1, -0.5, -1).normalize();
    scene.add(dirLight2);

    // Controls
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.1;

    // Add auth token to request
    const token = localStorage.getItem('access_token');
    const loader = new STLLoader();

    // Use XMLHttpRequest with auth header
    const xhr = new XMLHttpRequest();
    xhr.open('GET', fileUrl, true);
    xhr.responseType = 'arraybuffer';
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
    }

    xhr.onload = () => {
      if (xhr.status !== 200) {
        setError(`加载失败 (${xhr.status})`);
        setLoading(false);
        return;
      }
      try {
        const geometry = loader.parse(xhr.response);
        geometry.computeVertexNormals();

        const material = new THREE.MeshPhongMaterial({
          color: 0x8899aa,
          specular: 0x333333,
          shininess: 40,
          side: THREE.DoubleSide,
        });
        const mesh = new THREE.Mesh(geometry, material);
        scene.add(mesh);

        // Auto-center and fit
        geometry.computeBoundingBox();
        const box = geometry.boundingBox!;
        const center = new THREE.Vector3();
        box.getCenter(center);
        mesh.position.sub(center);

        const size = new THREE.Vector3();
        box.getSize(size);
        const maxDim = Math.max(size.x, size.y, size.z);
        const fov = camera.fov * (Math.PI / 180);
        const dist = maxDim / (2 * Math.tan(fov / 2)) * 1.5;

        camera.position.set(dist * 0.7, dist * 0.5, dist * 0.7);
        camera.lookAt(0, 0, 0);
        controls.target.set(0, 0, 0);
        controls.update();

        setLoading(false);
      } catch {
        setError('STL文件解析失败');
        setLoading(false);
      }
    };

    xhr.onerror = () => {
      setError('网络请求失败');
      setLoading(false);
    };

    xhr.send();

    // Animation loop
    let animId: number;
    const animate = () => {
      animId = requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    // Handle resize
    const onResize = () => {
      const w = container.clientWidth || 780;
      camera.aspect = w / height;
      camera.updateProjectionMatrix();
      renderer.setSize(w, height);
    };
    window.addEventListener('resize', onResize);

    // Cleanup
    cleanupRef.current = () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', onResize);
      controls.dispose();
      renderer.dispose();
      if (container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement);
      }
      scene.traverse((obj) => {
        if ((obj as THREE.Mesh).geometry) (obj as THREE.Mesh).geometry.dispose();
        if ((obj as THREE.Mesh).material) {
          const mat = (obj as THREE.Mesh).material;
          if (Array.isArray(mat)) mat.forEach(m => m.dispose());
          else (mat as THREE.Material).dispose();
        }
      });
    };

    return () => {
      cleanupRef.current?.();
      cleanupRef.current = null;
    };
  }, [open, fileUrl]);

  return (
    <Modal
      open={open}
      onCancel={onClose}
      title={`3D预览 - ${fileName}`}
      footer={null}
      width={830}
      destroyOnClose
    >
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 8 }} />}
      <div style={{ position: 'relative', minHeight: 500 }}>
        {loading && (
          <div style={{
            position: 'absolute', inset: 0, display: 'flex',
            alignItems: 'center', justifyContent: 'center', zIndex: 10,
            background: 'rgba(240,240,240,0.8)',
          }}>
            <Spin size="large" tip="加载3D模型..." />
          </div>
        )}
        <div ref={containerRef} style={{ width: '100%', height: 500 }} />
      </div>
    </Modal>
  );
};

export default STLViewer;

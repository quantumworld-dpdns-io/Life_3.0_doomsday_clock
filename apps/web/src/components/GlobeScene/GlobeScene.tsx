import { Canvas, useFrame } from "@react-three/fiber";
import { Stars } from "@react-three/drei";
import { useMemo, useRef } from "react";
import * as THREE from "three";

interface GlobeSceneProps {
  minutesToMidnight: number;
}

const vertexShader = `
  uniform float uTime;
  uniform float uCritical;
  varying vec3 vNormal;

  void main() {
    vNormal = normal;
    float breakup = sin(position.x * 18.0 + uTime) * cos(position.y * 14.0 - uTime);
    vec3 displaced = position + normal * breakup * 0.16 * uCritical;
    gl_Position = projectionMatrix * modelViewMatrix * vec4(displaced, 1.0);
  }
`;

const fragmentShader = `
  uniform vec3 uColor;
  varying vec3 vNormal;

  void main() {
    float fresnel = pow(1.0 - abs(dot(normalize(vNormal), vec3(0.0, 0.0, 1.0))), 2.0);
    vec3 color = mix(uColor * 0.42, uColor, fresnel + 0.35);
    gl_FragColor = vec4(color, 1.0);
  }
`;

const LowPolyGlobe = ({ minutesToMidnight }: GlobeSceneProps) => {
  const meshRef = useRef<THREE.Mesh>(null);
  const materialRef = useRef<THREE.ShaderMaterial>(null);
  const risk = THREE.MathUtils.clamp(1 - minutesToMidnight / 60, 0, 1);
  const color = useMemo(() => {
    const safe = new THREE.Color("#003300");
    const danger = new THREE.Color("#ff0000");
    return safe.lerp(danger, risk);
  }, [risk]);

  useFrame((state) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += 0.0018;
      meshRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.12) * 0.06;
    }
    if (materialRef.current) {
      materialRef.current.uniforms.uTime.value = state.clock.elapsedTime;
      materialRef.current.uniforms.uColor.value = color;
      materialRef.current.uniforms.uCritical.value = THREE.MathUtils.clamp((2 - minutesToMidnight) / 2, 0, 1);
    }
  });

  return (
    <mesh ref={meshRef}>
      <icosahedronGeometry args={[2.1, 2]} />
      <shaderMaterial
        ref={materialRef}
        vertexShader={vertexShader}
        fragmentShader={fragmentShader}
        uniforms={{
          uTime: { value: 0 },
          uCritical: { value: 0 },
          uColor: { value: color },
        }}
      />
      <lineSegments>
        <wireframeGeometry args={[new THREE.IcosahedronGeometry(2.105, 2)]} />
        <lineBasicMaterial color="#39ff14" transparent opacity={0.18} />
      </lineSegments>
    </mesh>
  );
};

export const GlobeScene = ({ minutesToMidnight }: GlobeSceneProps) => (
  <div className="globe-scene" aria-label="Low-poly risk globe">
    <Canvas camera={{ position: [0, 0.2, 5.2], fov: 48 }} dpr={[1, 1.8]}>
      <color attach="background" args={["#080909"]} />
      <ambientLight intensity={0.24} />
      <pointLight position={[4, 3, 3]} color="#ff2020" intensity={2.1} />
      <pointLight position={[-3, -1, 2]} color="#39ff14" intensity={0.8} />
      <Stars radius={45} depth={18} count={900} factor={2.2} fade speed={0.45} />
      <LowPolyGlobe minutesToMidnight={minutesToMidnight} />
    </Canvas>
  </div>
);

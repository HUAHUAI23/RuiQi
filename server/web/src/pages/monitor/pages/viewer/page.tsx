import React, { useEffect, useRef } from "react";
import { WebGLRenderer, PerspectiveCamera, Scene, AmbientLight, DirectionalLight, Color, MeshStandardMaterial, MeshBasicMaterial, DoubleSide, DirectionalLightHelper, PointLight, PointLightHelper, FrontSide, BackSide, MeshPhongMaterial, SphereGeometry, Mesh, ShaderChunk, ShaderMaterial } from "three";
import { OrbitControls } from "three/addons/controls/OrbitControls.js";
import ThreeGlobe from "three-globe";
import countries from "./files/globe-data-min.json";
import travelHistory from "./files/my-flights.json";
import { useState } from "react";
import type { FeatureCollection, GeoJsonProperties, Geometry } from "./geojson";

type FlightCollection = {
  type: string;
  order: number;
  from: string;
  to: string;
  flightCode: string;
  date: string;
  status: boolean;
  startLat: number;
  startLng: number;
  endLat: number;
  endLng: number;
  arcAlt: number;
}

const Globe = React.memo(() => {
  const globeRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<ReturnType<typeof requestAnimationFrame> | null>(null);
  const renderer = useRef(new WebGLRenderer({ antialias: true, logarithmicDepthBuffer: false, alpha: true }));
  const camera = useRef(new PerspectiveCamera(50, 1, 0.001, 1000));
  const scene = useRef(new Scene());
  const [width, setWidth] = useState(0);
  const [height, setHeight] = useState(0);

//   // 片元着色器
//   const fragmentShader = `
// ${ShaderChunk.logdepthbuf_pars_fragment}
// precision mediump float;
// void main() {
//   gl_FragColor = vec4(0, 1, 1, 0.7);
//   ${ShaderChunk.logdepthbuf_fragment}
// }
// `;
        
//   // 顶点着色器
//   const vertexShader = `
// ${ShaderChunk.logdepthbuf_pars_vertex}
// bool isPerspectiveMatrix(mat4) {
//   return true;
// }
// varying vec4 m_pos;
// void main () {
//   vec4 modelPosition = modelMatrix * vec4(position, 1.0);
//   gl_Position = projectionMatrix *  viewMatrix * modelPosition;
//   ${ShaderChunk.logdepthbuf_vertex}
// }
// `;

  
  const resizeObserver = new ResizeObserver((entries) => {
    if (!Array.isArray(entries) || !entries.length) return;
    for (const entry of entries) {
      setWidth(entry.contentRect.width);
      setHeight(entry.contentRect.height);
    }
  });

  const resize = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    if (resize.current !== null) clearTimeout(resize.current);
    resize.current = setTimeout(() => {
      resize.current = null;
      camera.current.aspect = width / height;
      camera.current.updateProjectionMatrix();
      renderer.current.setSize(width, height);
    }, 0);
  }, [width, height])

  useEffect(() => {
    if (!globeRef.current) return;
    const $el = globeRef.current
    resizeObserver.observe($el);
    return () => resizeObserver.unobserve($el);
  }, []);

  useEffect(() => {
    if (!globeRef.current) return;
    const el = globeRef.current;
    const rect = globeRef.current.getBoundingClientRect()
    setWidth(rect.width);
    setHeight(rect.height);
    
    renderer.current.setPixelRatio(window.devicePixelRatio);
    renderer.current.setSize(width, height);
    el.appendChild(renderer.current.domElement);
    el.style.setProperty("width", "100%");
    el.style.setProperty("height", "100%");
    el.style.setProperty("overflow", "hidden");

    const controls = new OrbitControls(camera.current, renderer.current.domElement);

    const animate = () => {
      camera.current.lookAt(scene.current.position);
      controls.update();
      renderer.current.render(scene.current, camera.current);
      animationRef.current = requestAnimationFrame(animate);

      return () => {
        if (animationRef.current !== null) cancelAnimationFrame(animationRef.current);
      }
    };

    const init = () => {
      /* Ambient Light */
      const ambientLight = new AmbientLight(new Color(0xFFFFFF), 1);
      /* Left Light */
      const dLight = new DirectionalLight(new Color(0x000000), 0.6);
      dLight.position.set(-400, 100, 400);
      /* Top Light Light */
      const dLight1 = new DirectionalLight(new Color(0xFFFFFF), 1);
      dLight1.position.set(-200, 500, 200);
      /* Point Light */
      const dLight2 = new PointLight(new Color(0xFFFFFF), 0.8);
      dLight2.position.set(-200, 500, 200);
      const dirLight = new DirectionalLight(0x000000, 1);
      dirLight.position.set(5, 3, 4);
      scene.current.add(ambientLight, dLight, dLight1, dLight2, dirLight);

      
      // const ambientLight = new AmbientLight(0xffffff, 1)
      // scene.current.add(ambientLight); // Ambient light

      // const pointLight = new PointLight(0xffffff, 1);
      // pointLight.position.set(5, 3, 4);
      // scene.current.add(pointLight);
      // scene.current.add(new PointLightHelper(pointLight, 50));
      // const dirLight = new DirectionalLight(0xffffff, 1);
      // dirLight.position.set(5, 3, 4);
      // scene.current.add(dirLight);
      // scene.current.add(new DirectionalLightHelper(dirLight, 10, new Color(0xFF0000)));

      camera.current.aspect = width / height;
      camera.current.position.z = 400;
      camera.current.position.x = 0;
      camera.current.position.y = 0;
      camera.current.updateProjectionMatrix();

      controls.enableDamping = true;
      // controls.dynamicDampingFactor = 0.01;
      controls.enablePan = false;
      controls.minDistance = 200;
      controls.maxDistance = 500;
      controls.rotateSpeed = 0.8;
      controls.zoomSpeed = 1;
      controls.autoRotate = false;
      controls.minPolarAngle = Math.PI / 5;
      controls.maxPolarAngle = Math.PI - Math.PI / 5;

      return () => {
        scene.current.remove(ambientLight, dLight, dLight1, dLight2, dirLight);
        ambientLight.dispose();
        dLight.dispose();
        dLight1.dispose();
        dLight2.dispose();
        dirLight.dispose();
        // scene.current.remove(pointLight);
        // pointLight.dispose();
        controls.dispose();
      }
    };

    const initGlobe = () => {
      const globe = new ThreeGlobe({ waitForGlobeReady: true, animateIn: true });
      scene.current.add(globe);
      // const globeMaterial = new MeshPhongMaterial({ color: new Color(0xa071da) });
      // const globeMaterial = new ShaderMaterial({ fragmentShader, vertexShader });
      const globeMaterial = new MeshStandardMaterial({ color: new Color(0xa071da), metalness: 1, roughness: 0.75, side: FrontSide, depthWrite: false });
      
      // const globe_material = new MeshStandardMaterial({ color: new Color(0xa071da), metalness: 1, roughness: 0.75, side: FrontSide, depthWrite: true });
      // const globe_geometry = new SphereGeometry(globe.getGlobeRadius()+10, 75, 75);
      // const globe_mesh = new Mesh(globe_geometry, globe_material);
      // globe_mesh.rotation.y = -Math.PI / 2;
      // scene.current.add(globe_mesh);

      globe
        .showGlobe(true)
        .globeMaterial(globeMaterial)
        .showAtmosphere(true) // 大气层
        .atmosphereColor("#a071da") // 大气层颜色
        .atmosphereAltitude(0.25); // 大气层高度

      globe
        .hexPolygonsData(countries.features) // 矩阵GeoJson
        .hexPolygonResolution(3) // 矩阵分辨率
        .hexPolygonAltitude(0.001) // 扩散
        .hexPolygonMargin(0.4) // 矩阵间距
        .hexPolygonColor((e) => ["KGZ", "KOR", "THA", "RUS", "UZB", "IDN", "KAZ", "MYS"].includes((e as FeatureCollection<Geometry, GeoJsonProperties>["features"][number]).properties!.ISO_A3) ? "rgba(255, 255, 255, 1)" : "rgba(241, 230, 255, 0.87)");

      globe.onGlobeReady(() => {
        const pointsData = Array.from(travelHistory.flights
          .reduce((p, c) => {
            p.set(`${c.startLat},${c.startLng}`, {
              size: 1,
              order: c.order,
              label: c.from || "",
              lat: c.startLat,
              lng: c.startLng,
            })
            p.set(`${c.endLat},${c.endLng}`, {
              size: 1,
              order: c.order,
              label: c.to || "",
              lat: c.endLat,
              lng: c.endLng,
            })
            return p;
          }, new Map<string, { size: number; order: number; label: string; lat: number; lng: number; }>()).values()
        );
        globe
          .arcsData(travelHistory.flights)
          .arcColor((e: unknown) => ((e as FlightCollection).status ? "#8ed4ff" : "#a071da"))
          .arcAltitude((e: unknown) => (e as FlightCollection).arcAlt)
          .arcStroke(0.3)
          .arcDashLength(0.9)
          .arcDashGap(4)
          .arcDashAnimateTime(1000)
          .arcsTransitionDuration(1000)
          .arcDashInitialGap((e: unknown) => (e as FlightCollection).order * 1);
        globe
          .ringsData(pointsData)
          .pointColor("#ffffaa")
          .pointsMerge(true)
          .pointRadius(0.25);
      })

      globe.rotateY(-Math.PI * (5 / 9));
      globe.rotateZ(-Math.PI / 6);

      return () => {
        scene.current.remove(globe);
      }
    };

    return ((...args: (() => void)[]) => () => args.forEach((dispose) => dispose()))(
      init(),
      initGlobe(),
      animate(),
      () => {
        el.removeChild(renderer.current.domElement)
        el.style.removeProperty("width");
        el.style.removeProperty("height");
        el.style.removeProperty("overflow");
      },
    );
  }, []);

  return <div ref={globeRef}></div>;
});

export default Globe;

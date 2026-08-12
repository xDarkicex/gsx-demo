// graph3d — a force-directed 3D rendering of the FOLLOWS graph.
// Reads the node/edge JSON from #graph-data and draws spheres +
// lines into #graph-canvas with three.js.
(function () {
  'use strict';

  var canvas = document.getElementById('graph-canvas');
  if (!canvas) return;

  // The graph data lives in a data attribute (script content is
  // raw text — entities aren't decoded there).
  var data = JSON.parse(canvas.getAttribute('data-graph') || '{"nodes":[],"edges":[]}');
  var nodes = data.nodes || [];
  var edges = data.edges || [];
  if (nodes.length === 0) return;

  // --- layout: few hundred spring iterations up front ---
  var pos = {};
  var vel = {};
  nodes.forEach(function (n) {
    pos[n.id] = {
      x: (Math.random() - 0.5) * 8,
      y: (Math.random() - 0.5) * 8,
      z: (Math.random() - 0.5) * 8,
    };
    vel[n.id] = { x: 0, y: 0, z: 0 };
  });

  var degree = {};
  nodes.forEach(function (n) { degree[n.id] = 0; });
  edges.forEach(function (e) {
    if (degree[e.from] !== undefined) degree[e.from]++;
    if (degree[e.to] !== undefined) degree[e.to]++;
  });

  var C = 40;   // repulsion
  var K = 0.06; // spring
  var dt = 0.08;

  function step() {
    nodes.forEach(function (a) {
      var pa = pos[a.id];
      nodes.forEach(function (b) {
        if (a.id === b.id) return;
        var pb = pos[b.id];
        var dx = pa.x - pb.x, dy = pa.y - pb.y, dz = pa.z - pb.z;
        var d2 = dx * dx + dy * dy + dz * dz + 1e-6;
        var f = C / d2;
        var d = Math.sqrt(d2);
        vel[a.id].x += (dx / d) * f;
        vel[a.id].y += (dy / d) * f;
        vel[a.id].z += (dz / d) * f;
      });
      // centering
      vel[a.id].x -= pa.x * 0.02;
      vel[a.id].y -= pa.y * 0.02;
      vel[a.id].z -= pa.z * 0.02;
    });
    edges.forEach(function (e) {
      if (!pos[e.from] || !pos[e.to]) return;
      var pa = pos[e.from], pb = pos[e.to];
      var dx = pb.x - pa.x, dy = pb.y - pa.y, dz = pb.z - pa.z;
      var d = Math.sqrt(dx * dx + dy * dy + dz * dz) || 1;
      var f = K * (d - 1.6);
      var fx = (dx / d) * f, fy = (dy / d) * f, fz = (dz / d) * f;
      vel[e.from].x += fx; vel[e.from].y += fy; vel[e.from].z += fz;
      vel[e.to].x -= fx; vel[e.to].y -= fy; vel[e.to].z -= fz;
    });
    nodes.forEach(function (n) {
      var v = vel[n.id];
      v.x *= 0.85; v.y *= 0.85; v.z *= 0.85;
      pos[n.id].x += v.x * dt;
      pos[n.id].y += v.y * dt;
      pos[n.id].z += v.z * dt;
    });
  }
  for (var i = 0; i < 220; i++) step();

  // --- scene ---
  var scene = new THREE.Scene();
  var camera = new THREE.PerspectiveCamera(55, canvas.clientWidth / canvas.clientHeight, 0.1, 100);
  camera.position.set(0, 2, 11);

  var renderer;
  try {
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  } catch (e) {
    canvas.innerHTML = '<div class="graph-fallback">The 3D view needs WebGL.</div>';
    return;
  }
  renderer.setSize(canvas.clientWidth, canvas.clientHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  canvas.appendChild(renderer.domElement);

  // ambient + a key light so the spheres read as 3D
  scene.add(new THREE.AmbientLight(0xffffff, 0.55));
  var key = new THREE.DirectionalLight(0xffffff, 0.9);
  key.position.set(4, 6, 8);
  scene.add(key);
  var rim = new THREE.DirectionalLight(0x38bdf8, 0.4);
  rim.position.set(-6, -4, -6);
  scene.add(rim);

  var maxDeg = 1;
  nodes.forEach(function (n) { if (degree[n.id] > maxDeg) maxDeg = degree[n.id]; });

  var byId = {};
  nodes.forEach(function (n) {
    byId[n.id] = n;
    var t = degree[n.id] / maxDeg; // 0..1 popularity
    var color = new THREE.Color().setHSL(0.55 - 0.28 * t, 0.85, 0.42 + 0.2 * t);
    var r = 0.28 + 0.14 * t;
    var mat = new THREE.MeshStandardMaterial({ color: color, roughness: 0.35, metalness: 0.1 });
    var mesh = new THREE.Mesh(new THREE.SphereGeometry(r, 24, 24), mat);
    mesh.position.set(pos[n.id].x, pos[n.id].y, pos[n.id].z);
    mesh.userData = { id: n.id, name: n.name };
    scene.add(mesh);
  });

  var lineMat = new THREE.LineBasicMaterial({
    color: 0x38bdf8,
    transparent: true,
    opacity: 0.35,
  });
  var lineGeo = new THREE.BufferGeometry();
  var verts = [];
  edges.forEach(function (e) {
    var a = pos[e.from], b = pos[e.to];
    if (!a || !b) return;
    verts.push(a.x, a.y, a.z, b.x, b.y, b.z);
  });
  lineGeo.setAttribute('position', new THREE.Float32BufferAttribute(verts, 3));
  scene.add(new THREE.LineSegments(lineGeo, lineMat));

  // --- interaction: minimal orbit (drag / wheel / auto-rotate) ---
  var theta = 0, phi = 0, dist = 11;
  var isDown = false, prev = { x: 0, y: 0 };
  canvas.addEventListener('mousedown', function (e) {
    isDown = true;
    prev = { x: e.clientX, y: e.clientY };
  });
  window.addEventListener('mouseup', function () { isDown = false; });
  window.addEventListener('mousemove', function (e) {
    if (!isDown) return;
    theta -= (e.clientX - prev.x) * 0.005;
    phi -= (e.clientY - prev.y) * 0.005;
    phi = Math.max(-1.3, Math.min(1.3, phi));
    prev = { x: e.clientX, y: e.clientY };
  });
  canvas.addEventListener('wheel', function (e) {
    e.preventDefault();
    dist = Math.max(4, Math.min(20, dist + e.deltaY * 0.01));
  }, { passive: false });

  // hover tooltip
  var tip = document.createElement('div');
  tip.className = 'graph-tooltip';
  tip.style.display = 'none';
  canvas.appendChild(tip);

  var raycaster = new THREE.Raycaster();
  var mouse = new THREE.Vector2();
  canvas.addEventListener('mousemove', function (ev) {
    var r = canvas.getBoundingClientRect();
    mouse.x = ((ev.clientX - r.left) / r.width) * 2 - 1;
    mouse.y = -((ev.clientY - r.top) / r.height) * 2 + 1;
    raycaster.setFromCamera(mouse, camera);
    var hits = raycaster.intersectObjects(scene.children.filter(function (o) { return o.isMesh; }));
    if (hits.length) {
      var u = hits[0].object.userData;
      tip.textContent = u.name + ' @' + u.id + ' · ' + degree[u.id] + ' connections';
      tip.style.display = 'block';
      tip.style.left = (ev.clientX - r.left + 12) + 'px';
      tip.style.top = (ev.clientY - r.top + 12) + 'px';
    } else {
      tip.style.display = 'none';
    }
  });
  canvas.addEventListener('mouseleave', function () { tip.style.display = 'none'; });

  function tick() {
    requestAnimationFrame(tick);
    if (!isDown) theta += 0.002; // slow auto-rotate when idle
    camera.position.set(
      dist * Math.sin(phi) * Math.cos(theta),
      dist * Math.cos(phi) * 0.6,
      dist * Math.sin(phi) * Math.sin(theta)
    );
    camera.lookAt(0, 0, 0);
    renderer.render(scene, camera);
  }
  tick();

  window.addEventListener('resize', function () {
    camera.aspect = canvas.clientWidth / canvas.clientHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(canvas.clientWidth, canvas.clientHeight);
  });
})();

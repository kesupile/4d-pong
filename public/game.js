let lastElement;
let body;

const log = (data) => {
  if (!body) {
    body = document.getElementsByTagName("body")[0];
  }

  const nextElement = document.createElement("pre");
  nextElement.textContent = JSON.stringify(data, null, 4);

  if (lastElement) {
    body.insertBefore(nextElement, lastElement);
  } else {
    body.appendChild(nextElement);
  }

  lastElement = nextElement;
  console.log(data);
};

window.addEventListener("DOMContentLoaded", () => {
  log("Starting...");
  const pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: "stun:stun.l.google.com:19302",
      },
    ],
  });

  pc.onnegotiationneeded = (e) =>
    pc
      .createOffer()
      .then((d) => pc.setLocalDescription(d))
      .catch(log);

  const sendChannel = pc.createDataChannel("foo");

  pc.oniceconnectionstatechange = (e) => log(pc.iceConnectionState);
  pc.onicecandidate = (event) => {
    if (event.candidate === null) {
      const sessionDescription = btoa(JSON.stringify(pc.localDescription));
      log("base64 session description", sessionDescription);

      fetch("/api/session-start", {
        method: "POST",
        body: JSON.stringify({ sessionDescription }),
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((res) => res.json())
        .then((v) => {
          log(v);
          const remoteDescription = JSON.parse(atob(v.sessionDescription));
          log(remoteDescription);

          pc.setRemoteDescription(remoteDescription);
        });
    }
  };
});

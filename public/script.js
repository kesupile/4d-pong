window.addEventListener("DOMContentLoaded", () => {
  const log = console.log;
  log("hello");

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
      console.log("base64 session description", sessionDescription);

      fetch("/api/session-start", {
        method: "POST",
        body: JSON.stringify({ sessionDescription }),
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((res) => res.json())
        .then((v) => {
          console.log(v);
          const remoteDescription = JSON.parse(atob(v.sessionDescription));
          console.log(remoteDescription);

          pc.setRemoteDescription(remoteDescription);
        });
    }
  };
});

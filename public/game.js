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

const dataChannelLabel = window.crypto.randomUUID();
let lastElement;
let body;

const getUrlInfo = (() => {
  let gameId;
  let origin;

  return () => {
    if (!gameId || !origin) {
      const [, , gameIdPart] = window.location.pathname.split("/");
      gameId = gameIdPart;

      origin = window.location.origin;

      log({ gameId, origin });
    }

    return { origin, gameId };
  };
})();

const getGameStatus = () => {
  const { origin, gameId } = getUrlInfo();
  return fetch(`${origin}/api/game/${gameId}/status`).then((r) => r.json());
};

const initialiseGame = (details) => {
  const gameContainer = document.getElementById("game");
  gameContainer.style.width = `${details.width}px`;
  gameContainer.style.height = `${details.height}px`;
  gameContainer.style.backgroundColor = "black";

  // const thisPlayerElement = document.createElement("div");
  // thisPlayerElement.id = "top";
  // thisPlayerElement.classList.add("player", "currentPlayer");
  // thisPlayerElement.style.width = `${message.playerDimensions[0]}px`;
  // thisPlayerElement.style.height = `${message.playerDimensions[1]}px`;
  // thisPlayerElement.style.left = `${message.playerCoordinates[0]}px`;
  // thisPlayerElement.style.top = `${message.playerCoordinates[1]}px`;
  // gameContainer.appendChild(thisPlayerElement);
};

const startNewConnection = () => {
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

  const sendChannel = pc.createDataChannel(dataChannelLabel);

  pc.oniceconnectionstatechange = (e) => log(pc.iceConnectionState);
  pc.onicecandidate = (event) => {
    if (event.candidate === null) {
      const sessionDescription = btoa(JSON.stringify(pc.localDescription));

      const { origin, gameId } = getUrlInfo();

      fetch(`${origin}/api/game/${gameId}/join`, {
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

  sendChannel.onmessage = (message) => {
    if (typeof message.data === "string") {
      const data = JSON.parse(message.data);
      log(data);

      switch (data.type) {
        case "init":
          initialiseGame(data);
        default:
          console.error("String message with unknown type");
      }
    }
  };
};

window.addEventListener("DOMContentLoaded", () => {
  log(`ID: ${dataChannelLabel}`);

  getGameStatus().then((res) => {
    if (res.acceptingConnections) {
      initialiseGame(res);
      return startNewConnection();
    }
    return console.error("Do something else here...");
  });
});

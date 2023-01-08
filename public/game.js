const debugOnScreen = false;

const log = (data) => {
  console.log(data);

  if (!debugOnScreen) {
    return;
  }

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

const game = {
  initialise(details) {
    const gameContainer = document.getElementById("game");
    gameContainer.style.width = `${details.width}px`;
    gameContainer.style.height = `${details.height}px`;
    gameContainer.style.backgroundColor = "black";

    this.containerElement = gameContainer;
    this.details = details;

    this.scale();
  },

  scale() {
    const container = this.getContainerElement();
    if (!container) {
      return;
    }

    const minHeightWidth = Math.min(window.innerHeight, window.innerWidth);
    const scaleFactor = minHeightWidth / this.details.width;

    container.style.transform = `scale(${scaleFactor})`;
  },

  getContainerElement() {
    return this.containerElement;
  },

  getPlayerElement(position) {
    const key = `${position}Player`;
    const element = this[key];

    if (element) {
      return element;
    }

    const thisPlayerElement = document.createElement("div");
    thisPlayerElement.style.width = `50px`;
    thisPlayerElement.style.height = `10px`;
    thisPlayerElement.classList.add("player");
    thisPlayerElement.id = key;
    this.getContainerElement().appendChild(thisPlayerElement);
    this[key] = thisPlayerElement;

    return this[key];
  },

  getActivePlayerElements() {
    const players = [];

    if (this.topPlayer) {
      players.push(this.topPlayer);
    }

    if (this.bottomPlayer) {
      players.push(this.bottomPlayer);
    }

    if (this.leftPlayer) {
      players.push(this.leftPlayer);
    }

    if (this.rightPlayer) {
      players.push(this.rightPlayer);
    }

    return players;
  },
};

const currentPlayer = {
  setPosition(position) {
    const oldPosition = this.position;
    this.position = position;

    if (oldPosition !== position) {
      game.getActivePlayerElements().forEach((element) => {
        const currentPlayerClass = "currentPlayer";
        if (element.id === `${position}Player`) {
          element.classList.add(currentPlayerClass);
        } else {
          element.classList.remove(currentPlayerClass);
        }
      });
    }
  },
  getElement() {
    return game.getPlayerElement(this.position);
  },
};

const drawPlayer = ({ position, isCurrent, x, y, height, width }) => {
  const playerElement = game.getPlayerElement(position);

  playerElement.style.left = `${x}px`;
  playerElement.style.top = `${y}px`;
  playerElement.style.width = `${width}px`;
  playerElement.style.height = `${height}px`;

  if (isCurrent) {
    currentPlayer.setPosition(position);
  }
};

const chunkSize = 6;
const positions = ["top", "bottom", "left", "right"];
const handleGeneralUpdateMessage = (data) => {
  let i = 1;
  for (const position of positions) {
    if (data[i]) {
      drawPlayer({
        position,
        isCurrent: data[i + 1],
        x: data[i + 2],
        width: data[i + 4],
        y: data[i + 3],
        height: data[i + 5],
      });
    }
    i += chunkSize;
  }
};

const handleMessage = (message) => {
  const data = new Uint8Array(message.data);
  switch (data[0]) {
    case 0:
      return handleGeneralUpdateMessage(data);
    default:
      return;
  }
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

  sendChannel.onmessage = handleMessage;
};

window.addEventListener("DOMContentLoaded", () => {
  log(`ID: ${dataChannelLabel}`);

  getGameStatus().then((res) => {
    if (res.acceptingConnections) {
      game.initialise(res);
      return startNewConnection();
    }
    return console.error("Do something else here...");
  });

  window.addEventListener("resize", () => game.scale());
});

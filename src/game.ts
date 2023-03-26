import {
  Game,
  GameElement,
  CurrentPlayer,
  Controls,
  KeyCodeByte,
  Position,
  ElementPosition,
  OriginatorData,
  PlayerStatus,
} from "./types";

const KEYBOARD_SEND_INTERVAL = 35;

let dataChannel: RTCDataChannel;

const dataChannelLabel = window.crypto.randomUUID();

const getUrlInfo = (() => {
  let gameId: string;
  let origin: string;

  return () => {
    if (!gameId || !origin) {
      const [, , gameIdPart] = window.location.pathname.split("/");
      gameId = gameIdPart;

      origin = window.location.origin;

      console.log({ gameId, origin });
    }

    return { origin, gameId };
  };
})();

interface NotifyUserOptions {
  persist?: boolean;
  sentiment?: "neutral" | "negative" | "positive";
  startFadeTimeout?: number;
  removeTimeout?: number;
}
const notifyUser = (text: string, options?: NotifyUserOptions) => {
  const {
    persist = false,
    sentiment = "neutral",
    startFadeTimeout = 1_000,
    removeTimeout = 2_100,
  } = options ?? {};
  const notificationContainer = document.querySelector(
    "#notificationContainer"
  )!;
  const notification = document.createElement("div");
  notification.classList.add("notification");
  if (sentiment !== "neutral") {
    notification.classList.add(sentiment);
  }

  notification.textContent = text;
  notificationContainer.appendChild(notification);

  if (persist) {
    return;
  }

  if (startFadeTimeout) {
    setTimeout(() => {
      notification.classList.add("fadeOut");
    }, startFadeTimeout);
  }

  if (removeTimeout) {
    setTimeout(() => {
      notificationContainer.removeChild(notification);
    }, removeTimeout);
  }
};

const getGameStatus = () => {
  const { origin, gameId } = getUrlInfo();
  return fetch(`${origin}/api/game/${gameId}/status`).then((r) => r.json());
};

const game: Game = {
  balls: [],
  details: {
    currentPlayerName: "",
    active: false,
    width: 0,
    height: 0,
    playerStatuses: {},
  },

  initialise({ playerStatuses, ...details }, currentPlayerName) {
    const gameContainer = document.getElementById("game")!;
    gameContainer.style.width = `${details.width}px`;
    gameContainer.style.height = `${details.height}px`;
    gameContainer.style.backgroundColor = "black";

    this.containerElement = gameContainer;
    this.details = { ...details, playerStatuses: {}, currentPlayerName };

    this.scale();
  },

  update({ playerStatuses, ...newDetails }) {
    const oldDetails = this.details;

    let playerStatusLookup: Record<string, PlayerStatus> =
      oldDetails.playerStatuses;

    const oldPlayerStatusCount = Object.keys(oldDetails.playerStatuses).length;
    if (playerStatuses.length > oldPlayerStatusCount) {
      playerStatuses.forEach((playerStatus) => {
        const name = playerStatus.name;
        if (
          !oldDetails.playerStatuses[name] &&
          name !== oldDetails.currentPlayerName
        ) {
          notifyUser(`${name} joined the game`);
        }
        playerStatusLookup[name] = playerStatus;
      });
    }

    if (this.details.active) {
      const playersLeft: string[] = [];
      let playersEjected = Object.keys(oldDetails.currentPlayerName);

      playerStatuses.forEach((playerStatus) => {
        const name = playerStatus.name;
        playersEjected = playersEjected.filter((player) => player !== name);
        playersLeft.push(name);
        playerStatusLookup[name] = playerStatus;
      });

      console.log({ playersEjected, playersLeft });
      playersEjected.forEach((player) => {
        notifyUser(`${player} left the game`, { sentiment: "negative" });
      });

      if (playersLeft.length === 1) {
        notifyUser("You win", { sentiment: "positive", persist: true });
      }
    }

    this.details = {
      ...oldDetails,
      ...newDetails,
      playerStatuses: playerStatusLookup,
    };
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
    const element = this.containerElement;
    if (!element) {
      throw new Error("Did you forget to create the container element?");
    }
    return element;
  },

  getPlayerKey(position) {
    return `${position}Player`;
  },

  getPlayerElement(position) {
    const key = this.getPlayerKey(position);
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

    return thisPlayerElement;
  },

  removePlayerElement(position) {
    const key = this.getPlayerKey(position);
    const player = this[key];
    if (!player) {
      return;
    }

    this.getContainerElement().removeChild(player);
    this[key] = undefined;
  },

  getActivePlayerElements() {
    const players: HTMLElement[] = [];

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

  getBall(index, radius) {
    let ball = this.balls[index];
    if (ball) {
      return ball;
    }

    ball = document.createElement("div");
    ball.dataset.ballId = String(index);
    ball.classList.add("ball");
    const diameter = radius * 2;
    ball.style.width = `${diameter}px`;
    ball.style.height = `${diameter}px`;

    this.getContainerElement().appendChild(ball);

    this.balls[index] = ball;
    return ball;
  },
};

const currentPlayer: CurrentPlayer = {
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

      this.setKeyboardEventListener();
    }
  },

  getPosition() {
    const position = this.position;
    if (!position) {
      throw new Error("Did you forget to set the player position?");
    }
    return position;
  },

  getElement() {
    return game.getPlayerElement(this.getPosition());
  },

  keyCodeByPositionMap: {
    top: {
      ArrowLeft: 0,
      ArrowRight: 1,
    },
    bottom: {
      ArrowLeft: 0,
      ArrowRight: 1,
    },
    left: {
      ArrowUp: 0,
      ArrowDown: 1,
    },
    right: {
      ArrowUp: 0,
      ArrowDown: 1,
    },
  },

  getValidKeyCodeByte(e) {
    const key = e.key as Controls;
    const position = this.getPosition();
    return this.keyCodeByPositionMap[position]![key]!;
  },

  setKeyboardEventListener() {
    window.removeEventListener("keydown", this.keydownEventListener!);
    window.removeEventListener("keyup", this.keyupEventListener!);

    let keyByte: KeyCodeByte | undefined;
    let moving = false;
    let sendKey: NodeJS.Timer;

    this.keydownEventListener = (e) => {
      const code = this.getValidKeyCodeByte(e);
      if (typeof code !== "number") {
        return;
      }

      if (keyByte !== code) {
        keyByte = code;
        clearInterval(sendKey);
      } else if (moving) {
        return;
      }

      const arr = new Uint8Array(2);
      arr[0] = 0;
      arr[1] = code;

      sendKey = setInterval(() => {
        dataChannel.send(arr);
      }, KEYBOARD_SEND_INTERVAL);

      console.log("setting moving...");
      moving = true;
    };

    this.keyupEventListener = (e) => {
      const code = this.getValidKeyCodeByte(e);
      if (typeof code !== "number") {
        return;
      }

      if (keyByte === code) {
        keyByte = undefined;
        clearInterval(sendKey);
        moving = false;
      }
    };

    window.addEventListener("keydown", this.keydownEventListener);
    window.addEventListener("keyup", this.keyupEventListener);
  },
};

interface DrawPlayerOptions extends GameElement {
  position: Position;
  isCurrent: number;
}

const transformElement = (element: HTMLElement, { x, y }: ElementPosition) => {
  element.style.transform = `translate(${x}px, ${y}px)`;
};

const drawPlayer = (options: DrawPlayerOptions) => {
  const { position, isCurrent, height, width, ...xy } = options;
  const playerElement = game.getPlayerElement(position);
  playerElement.style.width = `${width}px`;
  playerElement.style.height = `${height}px`;
  transformElement(playerElement, xy);

  if (isCurrent) {
    currentPlayer.setPosition(position);
  }
};

interface Ball extends ElementPosition {
  radius: number;
  index: number;
}
const drawBall = ({ radius, index, ...xy }: Ball) =>
  transformElement(game.getBall(index, radius), xy);

const playerChunkSize = 6;
const playerPositions: Position[] = ["top", "bottom", "left", "right"];
const ballChunkSize = 4;
const handleGeneralUpdateMessage = (data: Uint8Array) => {
  let i = 1;

  // Draw players
  for (const position of playerPositions) {
    if (data[i]) {
      drawPlayer({
        position,
        isCurrent: data[i + 1],
        x: data[i + 2],
        width: data[i + 4],
        y: data[i + 3],
        height: data[i + 5],
      });
    } else {
      game.removePlayerElement(position);
    }
    i += playerChunkSize;
  }

  // Draw balls
  const ballStartIndex = playerPositions.length * 6 - 1;
  while (i < data.length) {
    if (!data[i]) {
      return;
    }

    drawBall({
      index: Math.floor((i - ballStartIndex) / ballChunkSize),
      radius: data[i + 1],
      x: data[i + 2],
      y: data[i + 3],
    });

    i += ballChunkSize;
  }
};

const startGameCountdown = ([, timeLeft]: Uint8Array) => {
  notifyUser(`${timeLeft}`, {
    sentiment: "positive",
    removeTimeout: 800,
  });
};

const debounce = (fn: (...data: any[]) => any, timeout: number) => {
  let timer: NodeJS.Timeout;
  return (...data: Parameters<typeof fn>) => {
    clearTimeout(timer);
    timer = setTimeout(fn, timeout, ...data);
  };
};

const fetchAndUpdateGameStatus = debounce(
  () => getGameStatus().then((details) => game.update(details)),
  300
);

const handleMessage: RTCDataChannel["onmessage"] = (message) => {
  const data = new Uint8Array(message.data);

  switch (data[0]) {
    case 0:
      return handleGeneralUpdateMessage(data);
    case 1:
      return fetchAndUpdateGameStatus();
    case 2:
      return startGameCountdown(data);
    case 3:
      return fetchAndUpdateGameStatus();
    default:
      return;
  }
};

const startNewConnection = (playerName: string) => {
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
      .catch(console.log);

  dataChannel = pc.createDataChannel(dataChannelLabel);

  pc.oniceconnectionstatechange = (e) => console.log(pc.iceConnectionState);
  pc.onicecandidate = (event) => {
    if (event.candidate === null) {
      const sessionDescription = btoa(JSON.stringify(pc.localDescription));

      const { origin, gameId } = getUrlInfo();

      fetch(`${origin}/api/game/${gameId}/join`, {
        method: "POST",
        body: JSON.stringify({ sessionDescription, playerName }),
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

  dataChannel.onmessage = handleMessage;
};

const waitForPlayerDetails = () =>
  new Promise<string>((resolve) => {
    const introSection = document.getElementById("intro")!;
    introSection.classList.add("visible");

    const input = document.getElementById("playerName")!;
    input.addEventListener("keydown", (e) => {
      const value = (e.target as HTMLInputElement).value;
      if (e.code !== "Enter" || !value) {
        return;
      }

      introSection.classList.remove("visible");
      resolve(value);
    });
  });

const getGameData = () => {
  let storageData: OriginatorData | null = null;
  try {
    storageData = JSON.parse(
      sessionStorage.getItem("originator")!
    ) as OriginatorData;
  } catch (e) {}

  sessionStorage.clear();
  return storageData;
};

const getPlayerName = async () => {
  const gameData = getGameData();
  if (!gameData) {
    return waitForPlayerDetails();
  }
  return gameData.creatorName;
};

window.addEventListener("DOMContentLoaded", () => {
  getGameStatus().then(async (res) => {
    if (!res.acceptingConnections) {
      return alert("Cannot connect to game");
    }

    const playerName = await getPlayerName();
    game.initialise(res, playerName);
    startNewConnection(playerName);
  });

  window.addEventListener("resize", () => game.scale());
});

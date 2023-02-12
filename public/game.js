(function(global, factory) {
    if (typeof module === "object" && typeof module.exports === "object") factory(exports);
    else if (typeof define === "function" && define.amd) define([
        "exports"
    ], factory);
    else if (global = typeof globalThis !== "undefined" ? globalThis : global || self) factory(global.game = {});
})(this, function(exports) {
    "use strict";
    Object.defineProperty(exports, "__esModule", {
        value: true
    });
    function _arrayLikeToArray(arr, len) {
        if (len == null || len > arr.length) len = arr.length;
        for(var i = 0, arr2 = new Array(len); i < len; i++)arr2[i] = arr[i];
        return arr2;
    }
    function _arrayWithHoles(arr) {
        if (Array.isArray(arr)) return arr;
    }
    function _iterableToArrayLimit(arr, i) {
        var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"];
        if (_i == null) return;
        var _arr = [];
        var _n = true;
        var _d = false;
        var _s, _e;
        try {
            for(_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true){
                _arr.push(_s.value);
                if (i && _arr.length === i) break;
            }
        } catch (err) {
            _d = true;
            _e = err;
        } finally{
            try {
                if (!_n && _i["return"] != null) _i["return"]();
            } finally{
                if (_d) throw _e;
            }
        }
        return _arr;
    }
    function _nonIterableRest() {
        throw new TypeError("Invalid attempt to destructure non-iterable instance.\\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
    }
    function _slicedToArray(arr, i) {
        return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest();
    }
    function _unsupportedIterableToArray(o, minLen) {
        if (!o) return;
        if (typeof o === "string") return _arrayLikeToArray(o, minLen);
        var n = Object.prototype.toString.call(o).slice(8, -1);
        if (n === "Object" && o.constructor) n = o.constructor.name;
        if (n === "Map" || n === "Set") return Array.from(n);
        if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen);
    }
    var KEYBOARD_SEND_INTERVAL = 35;
    var dataChannel;
    var dataChannelLabel = window.crypto.randomUUID();
    var getUrlInfo = function() {
        var gameId;
        var origin;
        return function() {
            if (!gameId || !origin) {
                var _window_location_pathname_split = _slicedToArray(window.location.pathname.split("/"), 3), gameIdPart = _window_location_pathname_split[2];
                gameId = gameIdPart;
                origin = window.location.origin;
                console.log({
                    gameId: gameId,
                    origin: origin
                });
            }
            return {
                origin: origin,
                gameId: gameId
            };
        };
    }();
    var getGameStatus = function() {
        var _getUrlInfo = getUrlInfo(), origin = _getUrlInfo.origin, gameId = _getUrlInfo.gameId;
        return fetch("".concat(origin, "/api/game/").concat(gameId, "/status")).then(function(r) {
            return r.json();
        });
    };
    var game = {
        balls: [],
        details: {
            width: 0,
            height: 0
        },
        initialise: function initialise(details) {
            var gameContainer = document.getElementById("game");
            gameContainer.style.width = "".concat(details.width, "px");
            gameContainer.style.height = "".concat(details.height, "px");
            gameContainer.style.backgroundColor = "black";
            this.containerElement = gameContainer;
            this.details = details;
            this.scale();
        },
        scale: function scale() {
            var container = this.getContainerElement();
            if (!container) {
                return;
            }
            var minHeightWidth = Math.min(window.innerHeight, window.innerWidth);
            var scaleFactor = minHeightWidth / this.details.width;
            container.style.transform = "scale(".concat(scaleFactor, ")");
        },
        getContainerElement: function getContainerElement() {
            var element = this.containerElement;
            if (!element) {
                throw new Error("Did you forget to create the container element?");
            }
            return element;
        },
        getPlayerKey: function getPlayerKey(position) {
            return "".concat(position, "Player");
        },
        getPlayerElement: function getPlayerElement(position) {
            var key = this.getPlayerKey(position);
            var element = this[key];
            if (element) {
                return element;
            }
            var thisPlayerElement = document.createElement("div");
            thisPlayerElement.style.width = "50px";
            thisPlayerElement.style.height = "10px";
            thisPlayerElement.classList.add("player");
            thisPlayerElement.id = key;
            this.getContainerElement().appendChild(thisPlayerElement);
            this[key] = thisPlayerElement;
            return thisPlayerElement;
        },
        removePlayerElement: function removePlayerElement(position) {
            var key = this.getPlayerKey(position);
            var player = this[key];
            if (!player) {
                return;
            }
            this.getContainerElement().removeChild(player);
            this[key] = undefined;
        },
        getActivePlayerElements: function getActivePlayerElements() {
            var players = [];
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
        getBall: function getBall(index, radius) {
            var ball = this.balls[index];
            if (ball) {
                return ball;
            }
            ball = document.createElement("div");
            ball.dataset.ballId = String(index);
            ball.classList.add("ball");
            ball.style.width = "".concat(radius, "px");
            ball.style.height = "".concat(radius, "px");
            this.getContainerElement().appendChild(ball);
            this.balls[index] = ball;
            return ball;
        }
    };
    var currentPlayer = {
        setPosition: function setPosition(position) {
            var oldPosition = this.position;
            this.position = position;
            if (oldPosition !== position) {
                game.getActivePlayerElements().forEach(function(element) {
                    var currentPlayerClass = "currentPlayer";
                    if (element.id === "".concat(position, "Player")) {
                        element.classList.add(currentPlayerClass);
                    } else {
                        element.classList.remove(currentPlayerClass);
                    }
                });
                this.setKeyboardEventListener();
            }
        },
        getPosition: function getPosition() {
            var position = this.position;
            if (!position) {
                throw new Error("Did you forget to set the player position?");
            }
            return position;
        },
        getElement: function getElement() {
            return game.getPlayerElement(this.getPosition());
        },
        keyCodeByPositionMap: {
            top: {
                ArrowLeft: 0,
                ArrowRight: 1
            },
            bottom: {
                ArrowLeft: 0,
                ArrowRight: 1
            },
            left: {
                ArrowUp: 0,
                ArrowDown: 1
            },
            right: {
                ArrowUp: 0,
                ArrowDown: 1
            }
        },
        getValidKeyCodeByte: function getValidKeyCodeByte(e) {
            var key = e.key;
            var position = this.getPosition();
            return this.keyCodeByPositionMap[position][key];
        },
        setKeyboardEventListener: function setKeyboardEventListener() {
            var _this = this;
            window.removeEventListener("keydown", this.keydownEventListener);
            window.removeEventListener("keyup", this.keyupEventListener);
            var keyByte;
            var moving = false;
            var sendKey;
            this.keydownEventListener = function(e) {
                var code = _this.getValidKeyCodeByte(e);
                if (typeof code !== "number") {
                    return;
                }
                if (keyByte !== code) {
                    keyByte = code;
                    clearInterval(sendKey);
                } else if (moving) {
                    return;
                }
                var arr = new Uint8Array(2);
                arr[0] = 0;
                arr[1] = code;
                sendKey = setInterval(function() {
                    dataChannel.send(arr);
                }, KEYBOARD_SEND_INTERVAL);
                console.log("setting moving...");
                moving = true;
            };
            this.keyupEventListener = function(e) {
                var code = _this.getValidKeyCodeByte(e);
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
        }
    };
    var drawPlayer = function(options) {
        var position = options.position, isCurrent = options.isCurrent, x = options.x, y = options.y, height = options.height, width = options.width;
        var playerElement = game.getPlayerElement(position);
        playerElement.style.left = "".concat(x, "px");
        playerElement.style.top = "".concat(y, "px");
        playerElement.style.width = "".concat(width, "px");
        playerElement.style.height = "".concat(height, "px");
        if (isCurrent) {
            currentPlayer.setPosition(position);
        }
    };
    var drawBall = function(param) {
        var x = param.x, y = param.y, radius = param.radius, index = param.index;
        var ball = game.getBall(index, radius);
        ball.style.left = "".concat(x, "px");
        ball.style.top = "".concat(y, "px");
    };
    var playerChunkSize = 6;
    var playerPositions = [
        "top",
        "bottom",
        "left",
        "right"
    ];
    var ballChunkSize = 4;
    var handleGeneralUpdateMessage = function(data) {
        var i = 1;
        var _iteratorNormalCompletion = true, _didIteratorError = false, _iteratorError = undefined;
        try {
            // Draw players
            for(var _iterator = playerPositions[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true){
                var position = _step.value;
                if (data[i]) {
                    drawPlayer({
                        position: position,
                        isCurrent: data[i + 1],
                        x: data[i + 2],
                        width: data[i + 4],
                        y: data[i + 3],
                        height: data[i + 5]
                    });
                } else {
                    game.removePlayerElement(position);
                }
                i += playerChunkSize;
            }
        } catch (err) {
            _didIteratorError = true;
            _iteratorError = err;
        } finally{
            try {
                if (!_iteratorNormalCompletion && _iterator.return != null) {
                    _iterator.return();
                }
            } finally{
                if (_didIteratorError) {
                    throw _iteratorError;
                }
            }
        }
        // Draw balls
        var ballStartIndex = playerPositions.length * 6 - 1;
        while(i < data.length){
            if (!data[i]) {
                return;
            }
            drawBall({
                index: Math.floor((i - ballStartIndex) / ballChunkSize),
                radius: data[i + 1],
                x: data[i + 2],
                y: data[i + 3]
            });
            i += ballChunkSize;
        }
    };
    var handleMessage = function(message) {
        var data = new Uint8Array(message.data);
        switch(data[0]){
            case 0:
                return handleGeneralUpdateMessage(data);
            default:
                return;
        }
    };
    var startNewConnection = function() {
        var pc = new RTCPeerConnection({
            iceServers: [
                {
                    urls: "stun:stun.l.google.com:19302"
                }
            ]
        });
        pc.onnegotiationneeded = function(e) {
            return pc.createOffer().then(function(d) {
                return pc.setLocalDescription(d);
            }).catch(console.log);
        };
        dataChannel = pc.createDataChannel(dataChannelLabel);
        pc.oniceconnectionstatechange = function(e) {
            return console.log(pc.iceConnectionState);
        };
        pc.onicecandidate = function(event) {
            if (event.candidate === null) {
                var sessionDescription = btoa(JSON.stringify(pc.localDescription));
                var _getUrlInfo = getUrlInfo(), origin = _getUrlInfo.origin, gameId = _getUrlInfo.gameId;
                fetch("".concat(origin, "/api/game/").concat(gameId, "/join"), {
                    method: "POST",
                    body: JSON.stringify({
                        sessionDescription: sessionDescription
                    }),
                    headers: {
                        "Content-Type": "application/json"
                    }
                }).then(function(res) {
                    return res.json();
                }).then(function(v) {
                    console.log(v);
                    var remoteDescription = JSON.parse(atob(v.sessionDescription));
                    console.log(remoteDescription);
                    pc.setRemoteDescription(remoteDescription);
                });
            }
        };
        dataChannel.onmessage = handleMessage;
    };
    window.addEventListener("DOMContentLoaded", function() {
        console.log("ID: ".concat(dataChannelLabel));
        getGameStatus().then(function(res) {
            if (res.acceptingConnections) {
                game.initialise(res);
                return startNewConnection();
            }
            return console.error("Do something else here...");
        });
        window.addEventListener("resize", function() {
            return game.scale();
        });
    });
});

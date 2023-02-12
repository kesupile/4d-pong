(function(global, factory) {
    if (typeof module === "object" && typeof module.exports === "object") factory();
    else if (typeof define === "function" && define.amd) define([], factory);
    else if (global = typeof globalThis !== "undefined" ? globalThis : global || self) factory();
})(this, function() {
    "use strict";
    var handleNewGameButtonClick = function() {
        fetch("/api/new-game", {
            method: "POST",
            redirect: "follow"
        }).then(function(data) {
            return data.json();
        }).then(function(param) {
            var gameId = param.gameId;
            window.location.replace("/game/".concat(gameId));
        });
    };
    window.addEventListener("DOMContentLoaded", function() {
        var newGameButton = document.getElementById("new-game-btn");
        newGameButton.addEventListener("click", handleNewGameButtonClick);
    });
});

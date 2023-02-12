const handleNewGameButtonClick = () => {
  fetch("/api/new-game", {
    method: "POST",
    redirect: "follow",
  })
    .then((data) => data.json())
    .then(({ gameId }) => {
      window.location.replace(`/game/${gameId}`);
    });
};

window.addEventListener("DOMContentLoaded", () => {
  const newGameButton = document.getElementById("new-game-btn")!;
  newGameButton.addEventListener("click", handleNewGameButtonClick);
});

import { GameData, OriginatorData } from "./types";

const generateOriginatorData = (creatorName: string): OriginatorData => ({
  id: window.crypto.randomUUID(),
  creatorName: creatorName,
});

const submitForm = (data: GameData) => {
  fetch("/api/new-game", {
    method: "POST",
    body: JSON.stringify(data),
    redirect: "follow",
  })
    .then((data) => data.json())
    .then(({ gameId }) => {
      window.sessionStorage.setItem(
        "originator",
        JSON.stringify(generateOriginatorData(data.creatorName))
      );
      window.location.replace(`/game/${gameId}`);
    });
};

const buildJsonFormData = (form: HTMLFormElement) => {
  const data: Record<string, any> = {};
  for (const [key, value] of new FormData(form)) {
    data[key] = key === "nPlayers" ? Number(value) : value;
  }
  return data as GameData;
};

window.addEventListener("DOMContentLoaded", () => {
  const form = document.querySelector("form") as HTMLFormElement;
  form.addEventListener("submit", (e) => {
    e.preventDefault();
    submitForm(buildJsonFormData(form));
  });
});

export interface ElementPosition {
  x: number;
  y: number;
}

export interface GameElement extends ElementPosition {
  width: number;
  height: number;
}

export type GameDetails = Pick<GameElement, "height" | "width">;

export type Position = "top" | "bottom" | "left" | "right";

type PlayerPositionKey = `${Position}Player`;

type PlayerElementMap<G extends string = PlayerPositionKey> = {
  [K in G]?: HTMLElement;
};

export interface Game extends PlayerElementMap {
  balls: HTMLElement[];
  details: GameDetails;
  containerElement?: HTMLElement;
  getActivePlayerElements: () => HTMLElement[];
  getBall: (index: number, radius: number) => HTMLElement;
  getContainerElement: () => HTMLElement;
  getPlayerKey: (position: Position) => PlayerPositionKey;
  getPlayerElement: (position: Position) => HTMLElement;
  initialise: (details: GameDetails) => void;
  removePlayerElement: (position: Position) => void;
  scale: () => void;
}

export type Controls = "ArrowLeft" | "ArrowRight" | "ArrowUp" | "ArrowDown";

export type KeyCodeByte = 1 | 0;

export type KeyboardEventListener = (e: KeyboardEvent) => void;

export interface CurrentPlayer {
  keyCodeByPositionMap: Record<Position, { [K in Controls]?: KeyCodeByte }>;
  keydownEventListener?: KeyboardEventListener;
  keyupEventListener?: KeyboardEventListener;
  position?: Position;
  getElement: () => HTMLElement;
  getPosition: () => Position;
  getValidKeyCodeByte: (e: KeyboardEvent) => KeyCodeByte;
  setKeyboardEventListener: () => void;
  setPosition: (position: Position) => void;
}

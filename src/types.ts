export interface GameData {
  creatorName: string;
  nPlayers: number;
}

export interface OriginatorData {
  id: string;
  creatorName: string;
}

export interface ElementPosition {
  x: number;
  y: number;
}

export interface GameElement extends ElementPosition {
  width: number;
  height: number;
}

export type Position = "top" | "bottom" | "left" | "right";

export interface PlayerStatus {
  position: Position;
  name: string;
  isActive: boolean;
}

export interface GameDetails extends Pick<GameElement, "height" | "width"> {
  active: boolean;
  currentPlayerName: string;
  playerStatuses: Record<string, PlayerStatus>;
}

export interface APIGameDetails
  extends Omit<GameDetails, "currentPlayerName" | "playerStatuses"> {
  playerStatuses: PlayerStatus[];
}

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
  initialise: (details: APIGameDetails, playerName: string) => void;
  removePlayerElement: (position: Position) => void;
  scale: () => void;
  update: (details: APIGameDetails) => void;
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

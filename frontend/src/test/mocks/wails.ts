// Mock for @wailsio/runtime
export const Clipboard = {
  SetText: async () => true,
  GetText: async () => '',
};

export const EventsOn = () => {};
export const EventsOff = () => {};
export const EventsEmit = () => {};
export const WindowGetCurrent = () => ({});
export const WindowShow = () => {};
export const WindowHide = () => {};
export const WindowMaximise = () => {};
export const WindowToggleMaximise = () => {};
export const WindowUnmaximise = () => {};
export const WindowMinimise = () => {};
export const WindowSetSystemDefaultTitle = () => {};
export const WindowSetTitle = () => {};
export const WindowClose = () => {};
export const ScreenGetAll = () => [];
export const BrowserOpenURL = () => {};
export const Environment = {};

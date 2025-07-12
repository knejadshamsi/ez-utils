import type { FileEditMode } from '../store.svelte';

export interface ColorTheme {
  primary: [number, number, number, number];
  secondary: [number, number, number, number];
  highlight: [number, number, number, number];
  tentative: [number, number, number, number];
}

export const COLOR_THEMES: Record<FileEditMode, ColorTheme> = {
  POPULATION: {
    primary: [220, 38, 127, 200],      // Pink
    secondary: [220, 38, 127, 50],     // Pink with transparency
    highlight: [255, 255, 0, 150],     // Yellow
    tentative: [147, 51, 234, 150]    // Purple
  },
  NETWORK: {
    primary: [34, 197, 94, 200],       // Green
    secondary: [34, 197, 94, 50],      // Green with transparency
    highlight: [255, 255, 0, 150],     // Yellow
    tentative: [59, 130, 246, 150]    // Blue
  },
  PT: {
    primary: [59, 130, 246, 200],      // Blue
    secondary: [59, 130, 246, 50],     // Blue with transparency
    highlight: [255, 255, 0, 150],     // Yellow
    tentative: [168, 85, 247, 150]    // Purple
  }
};

export function getLayerStyles(fileEditMode: FileEditMode) {
  const theme = COLOR_THEMES[fileEditMode];
  
  return {
    point: {
      radius: 8,
      minRadius: 4,
      maxRadius: 20,
      outline: true,
      outlineWidth: 2
    },
    line: {
      width: 3,
      minWidth: 1,
      maxWidth: 10
    },
    polygon: {
      filled: true,
      stroked: true,
      lineWidth: 2
    },
    editHandle: {
      radius: 8,
      color: theme.highlight
    },
    tentative: {
      lineColor: theme.tentative,
      fillColor: [...theme.tentative.slice(0, 3), 50] as [number, number, number, number],
      lineWidth: 3
    }
  };
}
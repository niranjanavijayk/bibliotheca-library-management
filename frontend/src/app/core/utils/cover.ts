const PALETTE: [string, string][] = [
  ['#2f4a3a', '#4f7a5e'],
  ['#7a3b2e', '#c8704e'],
  ['#2d3e5a', '#5877a8'],
  ['#5a3a5c', '#9a6a9d'],
  ['#8a6a1f', '#d6b04e'],
  ['#1f4a4a', '#4a9a96'],
  ['#4a2f2f', '#a86a5a'],
  ['#3a3a5a', '#7a7ac0'],
];

function hash(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) | 0;
  return Math.abs(h);
}

/** Deterministic gradient so a book without artwork always gets the same cover. */
export function coverGradient(seed: string): string {
  const [a, b] = PALETTE[hash(seed) % PALETTE.length];
  return `linear-gradient(155deg, ${a}, ${b})`;
}

export function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0].toUpperCase())
    .join('');
}

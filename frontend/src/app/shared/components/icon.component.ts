import { Component, input } from '@angular/core';

export type IconName =
  | 'dashboard' | 'books' | 'members' | 'loans' | 'search' | 'plus' | 'edit' | 'trash' | 'eye'
  | 'return' | 'renew' | 'grid' | 'list' | 'alert' | 'check' | 'back' | 'menu' | 'close' | 'arrow' | 'sparkle';

/** Inline SVG icon set, switched with the @switch control-flow block. */
@Component({
  selector: 'app-icon',
  host: { '[style.width.px]': 'size()', '[style.height.px]': 'size()', style: 'display:inline-flex' },
  template: `
    <svg [attr.width]="size()" [attr.height]="size()" viewBox="0 0 24 24" fill="none" stroke="currentColor"
         stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      @switch (name()) {
        @case ('dashboard') { <rect x="3" y="3" width="7" height="9" rx="1.5"/><rect x="14" y="3" width="7" height="5" rx="1.5"/><rect x="14" y="12" width="7" height="9" rx="1.5"/><rect x="3" y="16" width="7" height="5" rx="1.5"/> }
        @case ('books') { <path d="M4 19V5a2 2 0 0 1 2-2h12v15H6a2 2 0 0 0-2 2z"/><path d="M4 19a2 2 0 0 0 2 2h12"/><path d="M9 7h5"/> }
        @case ('members') { <circle cx="9" cy="8" r="3.5"/><path d="M2.5 20c0-3.6 2.9-6 6.5-6s6.5 2.4 6.5 6"/><circle cx="17.5" cy="9" r="2.5"/><path d="M17 14.2c2.7.2 4.5 2.1 4.5 5"/> }
        @case ('loans') { <path d="M4 8h14l-4-4"/><path d="M20 16H6l4 4"/> }
        @case ('search') { <circle cx="11" cy="11" r="6.5"/><path d="m20 20-4.2-4.2"/> }
        @case ('plus') { <path d="M12 5v14M5 12h14"/> }
        @case ('edit') { <path d="M4 20h4L19 9l-4-4L4 16v4z"/> }
        @case ('trash') { <path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13"/> }
        @case ('eye') { <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/> }
        @case ('return') { <path d="M9 14 4 9l5-5"/><path d="M4 9h10a6 6 0 0 1 0 12h-3"/> }
        @case ('renew') { <path d="M20 12a8 8 0 1 1-2.5-5.8"/><path d="M20 4v5h-5"/> }
        @case ('grid') { <rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/> }
        @case ('list') { <path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/> }
        @case ('alert') { <path d="M12 3 2 20h20L12 3z"/><path d="M12 10v4M12 17h.01"/> }
        @case ('check') { <path d="m4 12 5 5L20 6"/> }
        @case ('back') { <path d="M19 12H5m6-6-6 6 6 6"/> }
        @case ('menu') { <path d="M4 6h16M4 12h16M4 18h16"/> }
        @case ('close') { <path d="M6 6l12 12M18 6 6 18"/> }
        @case ('arrow') { <path d="M5 12h14m-6-6 6 6-6 6"/> }
        @case ('sparkle') { <path d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5 18 18M18 6l-2.5 2.5M8.5 15.5 6 18"/> }
      }
    </svg>`,
})
export class IconComponent {
  readonly name = input.required<IconName>();
  readonly size = input(18);
}

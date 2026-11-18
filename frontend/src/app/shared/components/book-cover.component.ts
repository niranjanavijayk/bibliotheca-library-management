import { Component, computed, input, signal } from '@angular/core';

import { coverGradient } from '../../core/utils/cover';

/** Shows the cover image; falls back to a generated cover when the image is missing. */
@Component({
  selector: 'app-book-cover',
  host: { '[class]': '"cover cover--" + size()' },
  styleUrl: './book-cover.component.scss',
  template: `
    @if (showImage()) {
      <img class="cover__img" [src]="src()" [alt]="title() + ' cover'" loading="lazy"
           (load)="onLoad($event)" (error)="failedSrc.set(src())" />
    } @else {
      <div class="cover__fallback" [style.background]="gradient()">
        <span class="cover__ornament"></span>
        <span class="cover__title">{{ title() }}</span>
        <span class="cover__author">{{ author() }}</span>
      </div>
    }`,
})
export class BookCoverComponent {
  readonly title = input.required<string>();
  readonly author = input('');
  readonly src = input<string | null>('');
  readonly size = input<'xs' | 'sm' | 'md' | 'lg'>('md');

  protected readonly failedSrc = signal<string | null>(null);
  protected readonly showImage = computed(() => !!this.src() && this.failedSrc() !== this.src());
  protected readonly gradient = computed(() => coverGradient(this.title() + this.author()));

  protected onLoad(ev: Event): void {
    const img = ev.target as HTMLImageElement;
    if (img.naturalWidth <= 1) this.failedSrc.set(this.src()); // placeholder pixel from some CDNs
  }
}

import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';

import { Book } from '../../core/models/models';
import { BookCoverComponent } from '../../shared/components/book-cover.component';

/** Books standing on wooden shelves. Each shelf plank is drawn under every grid cell. */
@Component({
  selector: 'app-book-shelf',
  imports: [RouterLink, BookCoverComponent],
  styleUrl: './book-shelf.component.scss',
  template: `
    <div class="bookcase">
      @for (b of books(); track b.id) {
        <a class="slot" [routerLink]="['/app/books', b.id]" [attr.aria-label]="b.title + ' by ' + b.author">
          <div class="slot__stand">
            <app-book-cover class="slot__cover" [title]="b.title" [author]="b.author" [src]="b.coverUrl" size="md" />
            <span class="slot__tag" [class.slot__tag--out]="!b.available">{{ b.available ? b.availableCopies + ' in' : 'Out' }}</span>
          </div>
          <div class="slot__plank"></div>
          <span class="slot__title">{{ b.title }}</span>
          <span class="slot__author">{{ b.author }}</span>
        </a>
      }
    </div>`,
})
export class BookShelfComponent {
  readonly books = input.required<Book[]>();
}

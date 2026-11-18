import { Component, input } from '@angular/core';

@Component({
  selector: 'app-page-header',
  styleUrl: './page-header.component.scss',
  template: `
    <header class="page-header">
      <div>
        <h1 class="page-header__title">{{ title() }}</h1>
        @if (subtitle()) { <p class="page-header__subtitle">{{ subtitle() }}</p> }
      </div>
      <div class="page-header__actions"><ng-content /></div>
    </header>`,
})
export class PageHeaderComponent {
  readonly title = input.required<string>();
  readonly subtitle = input('');
}

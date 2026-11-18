import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { Dashboard } from '../../core/models/models';
import { DashboardService } from '../../core/services/dashboard.service';
import { BookCoverComponent } from '../../shared/components/book-cover.component';
import { IconComponent, IconName } from '../../shared/components/icon.component';

interface Feature {
  icon: IconName;
  title: string;
  text: string;
}

const cover = (isbn: string) => `https://covers.openlibrary.org/b/isbn/${isbn}-L.jpg?default=false`;

@Component({
  selector: 'app-landing',
  imports: [RouterLink, BookCoverComponent, IconComponent],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.scss',
})
export class LandingComponent implements OnInit {
  private readonly dashboardService = inject(DashboardService);

  /** Live numbers shown in the hero strip; stays hidden if the API is offline. */
  protected readonly stats = signal<Dashboard | null>(null);

  protected readonly heroBooks = [
    { title: 'Atomic Habits', author: 'James Clear', src: cover('9780735211292') },
    { title: 'The Alchemist', author: 'Paulo Coelho', src: cover('9780062315007') },
    { title: 'Dune', author: 'Frank Herbert', src: cover('9780441172719') },
    { title: 'The Hobbit', author: 'J.R.R. Tolkien', src: cover('9780547928227') },
    { title: 'Sapiens', author: 'Yuval Noah Harari', src: cover('9780062316110') },
  ];

  protected readonly spines = [
    { w: 14, h: 60, c: '#9a5a2e', plate: 'Annals', ornate: false, bands: 3 },
    { w: 17, h: 66, c: '#a8683a', plate: 'Ex Libris', ornate: true, bands: 2 },
    { w: 14, h: 63, c: '#8b4f2a', plate: '', ornate: false, bands: 2 },
    { w: 16, h: 80, c: '#9a5a30', plate: 'Atlas', ornate: true, bands: 3 },
    { w: 15, h: 68, c: '#5c2a35', plate: 'Verse', ornate: false, bands: 2 },
    { w: 14, h: 76, c: '#6a2c35', plate: '', ornate: false, bands: 0 },
    { w: 13, h: 62, c: '#c9b27c', plate: '', ornate: false, bands: 0 },
    { w: 15, h: 78, c: '#6b4630', plate: 'Chronicles', ornate: true, bands: 2 },
    { w: 14, h: 70, c: '#5a3a2a', plate: 'Opera', ornate: true, bands: 2 },
    { w: 14, h: 68, c: '#9a4f2a', plate: '', ornate: false, bands: 3 },
    { w: 16, h: 72, c: '#7a6a55', plate: '', ornate: false, bands: 0 },
    { w: 14, h: 66, c: '#b08a4e', plate: 'Poems', ornate: false, bands: 2 },
    { w: 12, h: 62, c: '#5a3a2e', plate: '', ornate: false, bands: 1 },
  ].map((s) => ({ ...s, bandList: Array.from({ length: s.bands }) }));

  protected readonly features: Feature[] = [
    { icon: 'books', title: 'A catalogue you can browse', text: 'Add, edit and organise books on a visual bookshelf. Search, filter by category and sort in a click.' },
    { icon: 'members', title: 'Members, made simple', text: 'Register readers, keep contact and membership details together, and see everything they have borrowed.' },
    { icon: 'loans', title: 'Issue and return in seconds', text: 'Unavailable books are never issued. Returns put copies straight back on the shelf.' },
    { icon: 'alert', title: 'Overdue, spotted early', text: 'Due dates are tracked automatically. Late loans are flagged with the days late and any fine.' },
    { icon: 'dashboard', title: 'A live dashboard', text: 'Books, members, loans and overdue counts update the moment something changes.' },
    { icon: 'sparkle', title: 'Thoughtful extras', text: 'One-click renewals, late fees, borrowing limits and full borrowing history for every book and member.' },
  ];

  protected readonly steps = [
    { n: '01', title: 'Stock the shelves', text: 'Add your books and register your members.' },
    { n: '02', title: 'Lend with confidence', text: 'Issue a book, set the due date and let Bibliotheca keep track.' },
    { n: '03', title: 'Welcome them back', text: 'Process returns, see history and keep the library running smoothly.' },
  ];

  ngOnInit(): void {
    this.dashboardService.stats().subscribe({ next: (s) => this.stats.set(s), error: () => this.stats.set(null) });
  }
}

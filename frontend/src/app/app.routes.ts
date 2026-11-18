import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    title: 'Bibliotheca · Your shelf, your story',
    loadComponent: () => import('./features/landing/landing.component').then((m) => m.LandingComponent),
  },
  {
    path: 'app',
    loadComponent: () => import('./layout/shell.component').then((m) => m.ShellComponent),
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
      {
        path: 'dashboard',
        title: 'Dashboard · Bibliotheca',
        loadComponent: () => import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
      },
      {
        path: 'books',
        title: 'Books · Bibliotheca',
        loadComponent: () => import('./features/books/books.component').then((m) => m.BooksComponent),
      },
      {
        path: 'books/:id',
        title: 'Book details · Bibliotheca',
        loadComponent: () => import('./features/books/book-detail.component').then((m) => m.BookDetailComponent),
      },
      {
        path: 'members',
        title: 'Members · Bibliotheca',
        loadComponent: () => import('./features/members/members.component').then((m) => m.MembersComponent),
      },
      {
        path: 'members/:id',
        title: 'Member details · Bibliotheca',
        loadComponent: () => import('./features/members/member-detail.component').then((m) => m.MemberDetailComponent),
      },
      {
        path: 'loans',
        title: 'Loans · Bibliotheca',
        loadComponent: () => import('./features/loans/loans.component').then((m) => m.LoansComponent),
      },
    ],
  },
  { path: '**', redirectTo: '' },
];

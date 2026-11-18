import { Component, signal } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { ConfirmHostComponent } from '../shared/components/confirm-host.component';
import { IconComponent, IconName } from '../shared/components/icon.component';
import { ToastContainerComponent } from '../shared/components/toast-container.component';

interface NavItem {
  path: string;
  label: string;
  icon: IconName;
}

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, RouterLink, RouterLinkActive, IconComponent, ToastContainerComponent, ConfirmHostComponent],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.scss',
})
export class ShellComponent {
  protected readonly menuOpen = signal(false);
  protected readonly nav: NavItem[] = [
    { path: '/app/dashboard', label: 'Dashboard', icon: 'dashboard' },
    { path: '/app/books', label: 'Books', icon: 'books' },
    { path: '/app/members', label: 'Members', icon: 'members' },
    { path: '/app/loans', label: 'Loans', icon: 'loans' },
  ];
}

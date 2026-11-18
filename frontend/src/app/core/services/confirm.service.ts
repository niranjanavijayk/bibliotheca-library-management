import { Injectable, signal } from '@angular/core';

export interface ConfirmOptions {
  title: string;
  message: string;
  confirmLabel?: string;
  danger?: boolean;
}

export interface ConfirmState extends ConfirmOptions {
  resolve: (ok: boolean) => void;
}

/** Promise-based confirmation dialog: `if (await confirm.ask({...})) { ... }` */
@Injectable({ providedIn: 'root' })
export class ConfirmService {
  readonly state = signal<ConfirmState | null>(null);

  ask(options: ConfirmOptions): Promise<boolean> {
    return new Promise((resolve) => {
      this.state.set({ ...options, resolve });
    });
  }

  answer(ok: boolean): void {
    this.state()?.resolve(ok);
    this.state.set(null);
  }
}

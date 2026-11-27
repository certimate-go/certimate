export interface ContactEmailsState {
  emails: string[];
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface ContactEmailsActions {
  fetchEmails: (refresh?: boolean) => Promise<string[]>;
  setEmails: (emails: string[]) => Promise<void>;
  addEmail: (email: string) => Promise<void>;
  removeEmail: (email: string) => Promise<void>;
}

export interface ContactEmailsStore extends ContactEmailsState, ContactEmailsActions {}

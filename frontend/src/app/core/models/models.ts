export interface Book {
  id: number;
  title: string;
  author: string;
  category: string;
  isbn: string;
  description: string;
  coverUrl: string;
  publishedYear: number;
  totalCopies: number;
  availableCopies: number;
  available: boolean;
  createdAt: string;
}

export type BookInput = Pick<
  Book,
  'title' | 'author' | 'category' | 'isbn' | 'description' | 'coverUrl' | 'publishedYear' | 'totalCopies'
>;

export type MemberStatus = 'active' | 'suspended';
export type MembershipType = 'Standard' | 'Premium' | 'Student';

export interface Member {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
  membershipType: MembershipType;
  status: MemberStatus;
  joinedAt: string;
  expiresAt: string;
  activeLoans: number;
}

export type MemberInput = Pick<Member, 'name' | 'email' | 'phone' | 'address' | 'membershipType' | 'status'>;

export type LoanStatus = 'issued' | 'overdue' | 'returned';

export interface Loan {
  id: number;
  bookId: number;
  memberId: number;
  bookTitle: string;
  bookAuthor: string;
  bookIsbn: string;
  bookCover: string;
  memberName: string;
  issuedAt: string;
  dueDate: string;
  returnedAt: string | null;
  fine: number;
  renewals: number;
  status: LoanStatus;
  daysOverdue: number;
}

export interface IssueInput {
  bookId: number;
  memberId: number;
  days: number;
}

export interface CategoryCount {
  name: string;
  count: number;
}

export interface Dashboard {
  totalBooks: number;
  totalCopies: number;
  availableCopies: number;
  totalMembers: number;
  issuedBooks: number;
  overdueBooks: number;
  categories: CategoryCount[];
  recentLoans: Loan[];
}

export interface Paged<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ApiError {
  status: number;
  code: string;
  message: string;
  fields?: Record<string, string>;
}

export type SortOrder = 'asc' | 'desc';

export interface BookQuery {
  q?: string;
  category?: string;
  available?: boolean | null;
  sort?: string;
  order?: SortOrder;
  page?: number;
  pageSize?: number;
}

export interface MemberQuery {
  q?: string;
  status?: string;
  sort?: string;
  order?: SortOrder;
  page?: number;
  pageSize?: number;
}

export type LoanTab = 'all' | 'active' | 'overdue' | 'returned';

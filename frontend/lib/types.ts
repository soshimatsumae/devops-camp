// バックエンドAPI(docs/API_SPEC.md)に対応する型定義。

export type TaskStatus = "todo" | "in_progress" | "done";

export interface Task {
  id: number;
  parent_id: number | null;
  list_id: number | null;
  title: string;
  description: string | null;
  status: TaskStatus;
  estimated_hours: number;
  actual_hours: number | null;
  due_date: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
  children?: Task[];
}

export interface TaskList {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface CalendarEntry {
  date: string;
  total_hours: number;
}

export interface PageMeta {
  total_count: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface Paginated<T> {
  data: T[];
  meta: PageMeta;
}

export interface LoginResponse {
  id: number;
  access_token: string;
  token_type: string;
  expires_in: number;
}

// sortクエリパラメータの許可値(バックエンドのvalidSortsホワイトリストと一致させる、ADR-0020)
export type TaskSort =
  | "due_date_asc"
  | "due_date_desc"
  | "estimated_hours_asc"
  | "estimated_hours_desc";

export interface TaskListParams {
  status?: TaskStatus;
  parent_id?: number | "null";
  list_id?: number | "null";
  q?: string;
  sort?: TaskSort;
  page?: number;
  per_page?: number;
}

import { getToken, clearToken } from "./auth";
import type {
  CalendarEntry,
  LoginResponse,
  Paginated,
  Task,
  TaskList,
  TaskListParams,
} from "./types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

// バックエンドのエラーレスポンス形式({ "error": { "code", "message" } }、internal/httpx/httpx.go参照)
export class ApiError extends Error {
  code: string;
  status: number;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

interface RequestOptions {
  method?: string;
  body?: unknown;
  query?: Record<string, string | number | undefined>;
  // ログイン・新規登録はAuthorizationヘッダーを付けず、401を通常のエラーとして
  // 呼び出し元に返す(セッション切れの強制リダイレクトとは区別する)。
  requiresAuth?: boolean;
}

function buildQuery(query?: Record<string, string | number | undefined>): string {
  if (!query) return "";
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.set(key, String(value));
  }
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, query, requiresAuth = true } = options;

  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";

  if (requiresAuth) {
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE_URL}${path}${buildQuery(query)}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) return undefined as T;

  const json = await res.json().catch(() => null);

  if (!res.ok) {
    const code: string = json?.error?.code ?? "UNKNOWN_ERROR";
    const message: string = json?.error?.message ?? "予期しないエラーが発生しました";

    // 認証必須のAPIで401が返ってきた場合は、未ログイン・トークン期限切れのどちらか
    // (フロントエンドからは区別できない、docs/assignment/step4.md「苦労したこと3.」)。
    // トークンを破棄してS-01へ強制リダイレクトする(docs/SCREEN_FLOW.md参照)。
    if (requiresAuth && res.status === 401) {
      clearToken();
      // lib/auth.tsのlogout()と同じ理由でハードリロードを使う。
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination
      if (typeof window !== "undefined") window.location.href = "/login";
    }

    throw new ApiError(res.status, code, message);
  }

  return json as T;
}

// ---- AUTH ----

export function register(input: {
  name: string;
  email: string;
  password: string;
}): Promise<{ id: number; name: string; email: string }> {
  return request("/auth/register", { method: "POST", body: input, requiresAuth: false });
}

export function login(input: {
  email: string;
  password: string;
}): Promise<LoginResponse> {
  return request("/auth/login", { method: "POST", body: input, requiresAuth: false });
}

// ---- TASKS ----

export function listTasks(params: TaskListParams = {}): Promise<Paginated<Task>> {
  return request("/tasks", { query: { ...params } });
}

export function getTask(id: number): Promise<Task> {
  return request(`/tasks/${id}`);
}

export interface CreateTaskInput {
  title: string;
  description?: string;
  parent_id?: number;
  list_id?: number;
  estimated_hours: number;
  due_date?: string;
}

export function createTask(input: CreateTaskInput): Promise<Task> {
  return request("/tasks", { method: "POST", body: input });
}

export interface UpdateTaskInput {
  title?: string;
  description?: string;
  status?: Task["status"];
  list_id?: number | null;
  actual_hours?: number;
  due_date?: string;
}

export function updateTask(id: number, input: UpdateTaskInput): Promise<Task> {
  return request(`/tasks/${id}`, { method: "PATCH", body: input });
}

export function deleteTask(id: number): Promise<void> {
  return request(`/tasks/${id}`, { method: "DELETE" });
}

export function getCalendar(params: { from?: string; to?: string } = {}): Promise<{
  data: CalendarEntry[];
}> {
  return request("/tasks/calendar", { query: { ...params } });
}

// ---- LISTS ----

export function listLists(): Promise<{ data: TaskList[] }> {
  return request("/lists");
}

export function createList(input: { name: string }): Promise<TaskList> {
  return request("/lists", { method: "POST", body: input });
}

export function updateList(id: number, input: { name: string }): Promise<TaskList> {
  return request(`/lists/${id}`, { method: "PATCH", body: input });
}

export function deleteList(id: number): Promise<void> {
  return request(`/lists/${id}`, { method: "DELETE" });
}

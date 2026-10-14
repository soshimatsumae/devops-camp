"use client";

import { X } from "lucide-react";
import { useEffect, useState } from "react";
import { ApiError, listLists, updateTask } from "@/lib/api";
import type { Task, TaskList, TaskStatus } from "@/lib/types";

// M-02: タスク編集モーダル。リスト選択欄は編集対象が親タスク(parent_id===null)のときだけ表示する。
export default function TaskEditModal({
  task,
  onClose,
  onUpdated,
}: {
  task: Task;
  onClose: () => void;
  onUpdated: () => void;
}) {
  const isParent = task.parent_id === null;
  const [lists, setLists] = useState<TaskList[]>([]);
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description ?? "");
  const [status, setStatus] = useState<TaskStatus>(task.status);
  const [actualHours, setActualHours] = useState(task.actual_hours?.toString() ?? "");
  const [dueDate, setDueDate] = useState(task.due_date ?? "");
  const [listId, setListId] = useState(task.list_id?.toString() ?? "");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isParent) return;
    listLists()
      .then((res) => setLists(res.data))
      .catch(() => {});
  }, [isParent]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await updateTask(task.id, {
        title,
        description: description || undefined,
        status,
        due_date: dueDate || undefined,
        actual_hours: status === "done" ? Number(actualHours) : undefined,
        list_id: isParent ? (listId ? Number(listId) : null) : undefined,
      });
      onUpdated();
    } catch (err) {
      if (err instanceof ApiError && err.code === "ACTUAL_HOURS_REQUIRED") {
        setError("実績の工数は必須です");
      } else if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setError("工数は0.5〜9999の範囲で入力してください");
      } else if (err instanceof ApiError && err.code === "INVALID_LIST_ID") {
        setError("選択したリストが見つかりません");
      } else {
        setError("タスクの更新に失敗しました");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 p-4">
      <div className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-zinc-900">タスクを編集</h3>
          <button
            type="button"
            onClick={onClose}
            className="flex items-center gap-1 text-xs text-zinc-400 hover:text-zinc-600"
          >
            <X className="h-3.5 w-3.5" />
            閉じる
          </button>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3 text-sm">
          <div>
            <label className="mb-1 block text-zinc-600">タイトル</label>
            <input
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            />
          </div>
          <div>
            <label className="mb-1 block text-zinc-600">詳細</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            />
          </div>
          <div>
            <label className="mb-1 block text-zinc-600">ステータス</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value as TaskStatus)}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            >
              <option value="todo">todo</option>
              <option value="in_progress">in_progress</option>
              <option value="done">done</option>
            </select>
          </div>
          {isParent && (
            <div>
              <label className="mb-1 block text-zinc-600">リスト</label>
              <select
                value={listId}
                onChange={(e) => setListId(e.target.value)}
                className="w-full rounded-md border border-zinc-300 px-3 py-2"
              >
                <option value="">未所属</option>
                {lists.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name}
                  </option>
                ))}
              </select>
            </div>
          )}
          {status === "done" && (
            <div>
              <label className="mb-1 block text-zinc-600">実績の工数(必須、時間)</label>
              <input
                required
                type="number"
                step={0.5}
                min={0.5}
                max={9999}
                value={actualHours}
                onChange={(e) => setActualHours(e.target.value)}
                className="w-full rounded-md border border-zinc-300 px-3 py-2"
              />
            </div>
          )}
          <div>
            <label className="mb-1 block text-zinc-600">締切日</label>
            <input
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            />
          </div>
          {error && (
            <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-red-700">
              {error}
            </p>
          )}
          <div className="mt-1 flex gap-2">
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 rounded-md bg-sky-700 px-4 py-2 font-semibold text-white hover:bg-sky-800 disabled:opacity-60"
            >
              {submitting ? "保存中..." : "保存"}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-dashed border-zinc-300 px-4 py-2 text-zinc-600"
            >
              キャンセル
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

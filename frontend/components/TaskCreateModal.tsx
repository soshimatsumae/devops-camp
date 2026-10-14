"use client";

import { X } from "lucide-react";
import { useEffect, useState } from "react";
import { ApiError, createTask, listLists } from "@/lib/api";
import type { TaskList } from "@/lib/types";

// M-01: タスク作成モーダル(docs/SCREEN_LIST.md参照)。
// 現時点ではS-03の「+ 新規タスク作成」からの親タスク作成のみ対応し、
// M-03の「子タスクを追加」経由(parent_idの引き継ぎ)はM-03実装時に追加する。
export default function TaskCreateModal({
  parentId,
  onClose,
  onCreated,
}: {
  parentId?: number;
  onClose: () => void;
  onCreated: () => void;
}) {
  const isChild = parentId !== undefined;
  const [lists, setLists] = useState<TaskList[]>([]);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [estimatedHours, setEstimatedHours] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [listId, setListId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    // 子タスクはlist_idを持てないため、親タスク作成時のみリスト選択肢を取得する(ADR-0021)
    if (isChild) return;
    listLists()
      .then((res) => setLists(res.data))
      .catch(() => {
        // リスト選択肢が取得できなくても、未選択のまま作成は続行できる
      });
  }, [isChild]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await createTask({
        title,
        description: description || undefined,
        estimated_hours: Number(estimatedHours),
        due_date: dueDate || undefined,
        parent_id: parentId,
        list_id: isChild ? undefined : listId ? Number(listId) : undefined,
      });
      onCreated();
    } catch (err) {
      if (err instanceof ApiError && err.code === "ESTIMATED_HOURS_REQUIRED") {
        setError("想定の工数は必須です");
      } else if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setError("工数は0.5〜9999の範囲で入力してください");
      } else if (err instanceof ApiError && err.code === "INVALID_LIST_ID") {
        setError("選択したリストが見つかりません");
      } else {
        setError("タスクの作成に失敗しました");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 p-4">
      <div className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-zinc-900">新規タスク作成</h3>
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
          {isChild && (
            <p className="rounded-md border border-dashed border-zinc-300 px-3 py-2 text-xs text-zinc-500">
              子タスクとして作成します
            </p>
          )}
          <div>
            <label className="mb-1 block text-zinc-600">タイトル(必須)</label>
            <input
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            />
          </div>
          {!isChild && (
            <div>
              <label className="mb-1 block text-zinc-600">リスト(任意)</label>
              <select
                value={listId}
                onChange={(e) => setListId(e.target.value)}
                className="w-full rounded-md border border-zinc-300 px-3 py-2"
              >
                <option value="">未選択</option>
                {lists.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name}
                  </option>
                ))}
              </select>
            </div>
          )}
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
            <label className="mb-1 block text-zinc-600">想定の工数(必須、時間)</label>
            <input
              required
              type="number"
              step={0.5}
              min={0.5}
              max={9999}
              value={estimatedHours}
              onChange={(e) => setEstimatedHours(e.target.value)}
              className="w-full rounded-md border border-zinc-300 px-3 py-2"
            />
          </div>
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
              {submitting ? "作成中..." : "作成"}
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

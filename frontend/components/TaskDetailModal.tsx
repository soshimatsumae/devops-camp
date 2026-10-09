"use client";

import { useEffect, useState } from "react";
import { ApiError, deleteTask, getTask, listTasks } from "@/lib/api";
import type { Task } from "@/lib/types";
import TaskCreateModal from "./TaskCreateModal";
import TaskEditModal from "./TaskEditModal";
import TaskTree from "./TaskTree";

// M-03: タスク詳細モーダル。子タスク一覧はTaskTreeで孫タスク以下もインライン展開できる。
export default function TaskDetailModal({
  taskId,
  onClose,
  onChanged,
}: {
  taskId: number;
  onClose: () => void;
  onChanged: () => void;
}) {
  const [task, setTask] = useState<Task | null>(null);
  const [children, setChildren] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [childCreateOpen, setChildCreateOpen] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  function load() {
    setLoading(true);
    setError(null);
    Promise.all([getTask(taskId), listTasks({ parent_id: taskId })])
      .then(([t, childRes]) => {
        setTask(t);
        setChildren(childRes.data);
      })
      .catch(() => setError("タスク詳細の取得に失敗しました"))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [taskId]);

  async function handleDelete() {
    if (!window.confirm("このタスクを削除しますか？子・孫タスクも連動削除されます")) return;
    setDeleteError(null);
    try {
      await deleteTask(taskId);
      onChanged();
      onClose();
    } catch (err) {
      if (err instanceof ApiError && err.code === "TASK_NOT_FOUND") {
        setDeleteError("既に削除済み、または他ユーザーのタスクです");
      } else {
        setDeleteError("削除に失敗しました");
      }
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 p-4">
      <div className="max-h-[90vh] w-full max-w-md overflow-y-auto rounded-xl border border-zinc-200 bg-white p-6 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-zinc-900">
            {loading ? "読み込み中..." : task?.title}
            {task && (
              <span className="ml-2 rounded border border-zinc-300 px-1.5 py-0.5 align-middle text-xs font-normal text-zinc-500">
                {task.status}
              </span>
            )}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-xs text-zinc-400 hover:text-zinc-600"
          >
            ✕ 閉じる
          </button>
        </div>

        {error && (
          <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        )}

        {task && (
          <div className="flex flex-col gap-4 text-sm">
            {task.description && (
              <p className="rounded-md border border-zinc-200 px-3 py-2 text-zinc-700">
                {task.description}
              </p>
            )}

            <div className="grid grid-cols-2 gap-x-4 gap-y-2 font-mono text-xs text-zinc-500">
              <div>
                <span className="block">ESTIMATED_HOURS</span>
                <span className="font-sans text-sm text-zinc-900">
                  {task.estimated_hours}時間
                </span>
              </div>
              <div>
                <span className="block">ACTUAL_HOURS</span>
                <span className="font-sans text-sm text-zinc-900">
                  {task.actual_hours ?? "-"}
                </span>
              </div>
              <div>
                <span className="block">DUE_DATE</span>
                <span className="font-sans text-sm text-zinc-900">
                  {task.due_date ?? "-"}
                </span>
              </div>
              <div>
                <span className="block">COMPLETED_AT</span>
                <span className="font-sans text-sm text-zinc-900">
                  {task.completed_at ?? "-"}
                </span>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => setEditOpen(true)}
                className="rounded-md border border-zinc-300 px-3 py-1.5 text-zinc-700 hover:bg-zinc-50"
              >
                編集
              </button>
              <button
                type="button"
                onClick={() => setChildCreateOpen(true)}
                className="rounded-md border border-zinc-300 px-3 py-1.5 text-zinc-700 hover:bg-zinc-50"
              >
                子タスクを追加
              </button>
              <button
                type="button"
                onClick={handleDelete}
                className="rounded-md border border-red-300 px-3 py-1.5 text-red-700 hover:bg-red-50"
              >
                削除
              </button>
            </div>

            {deleteError && (
              <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-red-700">
                {deleteError}
              </p>
            )}

            <div>
              <h4 className="mb-2 text-xs font-semibold text-zinc-500">子タスク</h4>
              {children.length === 0 ? (
                <p className="text-xs text-zinc-400">子タスクなし</p>
              ) : (
                <div className="flex flex-col gap-2">
                  {children.map((c) => (
                    <TaskTree key={c.id} task={c} mode="modal" />
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {editOpen && task && (
        <TaskEditModal
          task={task}
          onClose={() => setEditOpen(false)}
          onUpdated={() => {
            setEditOpen(false);
            load();
            onChanged();
          }}
        />
      )}

      {childCreateOpen && task && (
        <TaskCreateModal
          parentId={task.id}
          onClose={() => setChildCreateOpen(false)}
          onCreated={() => {
            setChildCreateOpen(false);
            load();
            onChanged();
          }}
        />
      )}
    </div>
  );
}

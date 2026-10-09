"use client";

import { useEffect, useState } from "react";
import { ApiError, createList, deleteList, listLists, updateList } from "@/lib/api";
import type { TaskList } from "@/lib/types";

// M-04: リスト管理モーダル。作成・改名・削除をここでまとめて行う。
export default function ListManageModal({
  onClose,
  onChanged,
}: {
  onClose: () => void;
  onChanged: () => void;
}) {
  const [lists, setLists] = useState<TaskList[]>([]);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");
  const [rowError, setRowError] = useState<string | null>(null);

  function load() {
    setLoading(true);
    listLists()
      .then((res) => setLists(res.data))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreateError(null);
    try {
      await createList({ name: newName });
      setNewName("");
      load();
      onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setCreateError("リスト名を入力してください");
      } else {
        setCreateError("作成に失敗しました");
      }
    }
  }

  async function handleRename(id: number) {
    setRowError(null);
    try {
      await updateList(id, { name: editingName });
      setEditingId(null);
      load();
      onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setRowError("リスト名を入力してください");
      } else {
        setRowError("更新に失敗しました");
      }
    }
  }

  async function handleDelete(id: number) {
    if (!window.confirm("このリストを削除しますか？所属するタスクは削除されず、未所属に戻ります")) {
      return;
    }
    setRowError(null);
    try {
      await deleteList(id);
      load();
      onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.code === "LIST_NOT_FOUND") {
        setRowError("既に削除済み、または他ユーザーのリストです");
      } else {
        setRowError("削除に失敗しました");
      }
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 p-4">
      <div className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-zinc-900">リストを管理</h3>
          <button
            type="button"
            onClick={onClose}
            className="text-xs text-zinc-400 hover:text-zinc-600"
          >
            ✕ 閉じる
          </button>
        </div>

        <form onSubmit={handleCreate} className="mb-4 flex gap-2 text-sm">
          <input
            required
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="新しいリスト名"
            className="flex-1 rounded-md border border-zinc-300 px-3 py-2"
          />
          <button
            type="submit"
            className="rounded-md bg-sky-700 px-3 py-2 text-sm font-semibold text-white hover:bg-sky-800"
          >
            + 追加
          </button>
        </form>
        {createError && (
          <p className="mb-3 rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
            {createError}
          </p>
        )}

        {loading ? (
          <p className="text-sm text-zinc-400">読み込み中...</p>
        ) : lists.length === 0 ? (
          <p className="text-sm text-zinc-400">リストはまだありません</p>
        ) : (
          <div className="flex flex-col gap-2">
            {lists.map((l) => (
              <div
                key={l.id}
                className="flex items-center justify-between gap-2 rounded-md border border-dashed border-zinc-300 px-3 py-2 text-sm"
              >
                {editingId === l.id ? (
                  <>
                    <input
                      autoFocus
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      className="flex-1 rounded-md border border-zinc-300 px-2 py-1"
                    />
                    <div className="flex gap-2 whitespace-nowrap text-xs">
                      <button
                        type="button"
                        onClick={() => handleRename(l.id)}
                        className="text-sky-700 underline"
                      >
                        保存
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingId(null)}
                        className="text-zinc-500 underline"
                      >
                        キャンセル
                      </button>
                    </div>
                  </>
                ) : (
                  <>
                    <span>{l.name}</span>
                    <div className="flex gap-2 whitespace-nowrap text-xs">
                      <button
                        type="button"
                        onClick={() => {
                          setEditingId(l.id);
                          setEditingName(l.name);
                        }}
                        className="text-sky-700 underline"
                      >
                        ✎ 編集
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(l.id)}
                        className="text-red-700 underline"
                      >
                        🗑 削除
                      </button>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}

        {rowError && (
          <p className="mt-3 rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
            {rowError}
          </p>
        )}
      </div>
    </div>
  );
}

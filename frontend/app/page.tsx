"use client";

import { useState } from "react";
import AppHeader from "@/components/AppHeader";
import Heatmap from "@/components/Heatmap";
import RequireAuth from "@/components/RequireAuth";
import TaskCreateModal from "@/components/TaskCreateModal";
import TaskDetailModal from "@/components/TaskDetailModal";
import TaskListSection from "@/components/TaskListSection";

export default function DashboardPage() {
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [openTaskId, setOpenTaskId] = useState<number | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  return (
    <RequireAuth>
      <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-4 p-4">
        <AppHeader />
        <Heatmap />
        <TaskListSection
          onCreateClick={() => setCreateModalOpen(true)}
          onOpenTask={(id) => setOpenTaskId(id)}
          refreshKey={refreshKey}
        />
      </div>

      {createModalOpen && (
        <TaskCreateModal
          onClose={() => setCreateModalOpen(false)}
          onCreated={() => {
            setCreateModalOpen(false);
            setRefreshKey((k) => k + 1);
          }}
        />
      )}

      {openTaskId !== null && (
        <TaskDetailModal
          taskId={openTaskId}
          onClose={() => setOpenTaskId(null)}
          onChanged={() => setRefreshKey((k) => k + 1)}
        />
      )}
    </RequireAuth>
  );
}

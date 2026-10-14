import type { Task } from "./types";

export const STATUS_LABEL: Record<Task["status"], string> = {
  todo: "todo",
  in_progress: "in_progress",
  done: "done",
};

export const STATUS_CHIP_CLASS: Record<Task["status"], string> = {
  todo: "border-zinc-300 bg-zinc-100 text-zinc-600",
  in_progress: "border-amber-300 bg-amber-100 text-amber-700",
  done: "border-emerald-300 bg-emerald-100 text-emerald-700",
};

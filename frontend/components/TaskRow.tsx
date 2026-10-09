import type { Task } from "@/lib/types";

const STATUS_LABEL: Record<Task["status"], string> = {
  todo: "todo",
  in_progress: "in_progress",
  done: "done",
};

export default function TaskRow({
  task,
  onClick,
  caret,
  onCaretClick,
  depth = 0,
}: {
  task: Task;
  onClick?: () => void;
  caret?: "expanded" | "collapsed";
  onCaretClick?: () => void;
  depth?: number;
}) {
  return (
    <div
      onClick={onClick}
      style={{ marginLeft: depth * 28 }}
      className={`flex items-center justify-between gap-2 rounded-md border border-zinc-200 px-3 py-2 ${
        onClick ? "cursor-pointer hover:bg-zinc-50" : ""
      }`}
    >
      <span className="flex items-center gap-2">
        {caret && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onCaretClick?.();
            }}
            className="w-4 shrink-0 font-mono text-xs text-sky-700"
          >
            {caret === "expanded" ? "▾" : "▸"}
          </button>
        )}
        <span className="text-[0.95rem] text-zinc-900">{task.title}</span>
      </span>
      <span className="flex items-center gap-2 font-mono text-xs text-zinc-500">
        <span
          className={`rounded border px-1.5 py-0.5 ${
            task.status === "done"
              ? "border-sky-500 text-sky-700"
              : "border-zinc-300"
          }`}
        >
          {STATUS_LABEL[task.status]}
        </span>
        <span>
          想定:{task.estimated_hours}h 実績:{task.actual_hours ?? "-"}
        </span>
      </span>
    </div>
  );
}

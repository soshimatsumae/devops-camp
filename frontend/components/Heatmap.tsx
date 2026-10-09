"use client";

import { useEffect, useState } from "react";
import { ApiError, getCalendar } from "@/lib/api";
import { defaultCalendarRange, enumerateDates } from "@/lib/calendarRange";

function levelFor(hours: number): 0 | 1 | 2 | 3 | 4 {
  if (hours <= 0) return 0;
  if (hours <= 2) return 1;
  if (hours <= 4) return 2;
  if (hours <= 8) return 3;
  return 4;
}

const LEVEL_CLASS: Record<number, string> = {
  0: "bg-zinc-100",
  1: "bg-sky-100",
  2: "bg-sky-300",
  3: "bg-sky-500",
  4: "bg-sky-700",
};

export default function Heatmap() {
  const [range, setRange] = useState(() => defaultCalendarRange());
  const [fromInput, setFromInput] = useState(range.from);
  const [toInput, setToInput] = useState(range.to);
  const [hoursByDate, setHoursByDate] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    getCalendar({ from: range.from, to: range.to })
      .then((res) => {
        if (cancelled) return;
        const map: Record<string, number> = {};
        for (const entry of res.data) map[entry.date] = entry.total_hours;
        setHoursByDate(map);
      })
      .catch((err) => {
        if (cancelled) return;
        if (err instanceof ApiError && err.code === "INVALID_DATE_FORMAT") {
          setError("開始日・終了日はYYYY-MM-DD形式で入力してください");
        } else if (err instanceof ApiError && err.code === "INVALID_DATE_RANGE") {
          setError("開始日は終了日より前の日付にしてください");
        } else {
          setError("ヒートマップの取得に失敗しました");
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [range]);

  function handleApply(e: React.FormEvent) {
    e.preventDefault();
    setRange({ from: fromInput, to: toInput });
  }

  const dates = enumerateDates(range.from, range.to);
  const weekCount = Math.ceil(dates.length / 7);

  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-zinc-500">頑張りの可視化</h2>
      </div>

      <form onSubmit={handleApply} className="flex flex-wrap items-center gap-2 text-sm">
        <label className="flex items-center gap-1 text-zinc-600">
          開始日
          <input
            type="date"
            value={fromInput}
            onChange={(e) => setFromInput(e.target.value)}
            className="rounded-md border border-zinc-300 px-2 py-1"
          />
        </label>
        <label className="flex items-center gap-1 text-zinc-600">
          終了日
          <input
            type="date"
            value={toInput}
            onChange={(e) => setToInput(e.target.value)}
            className="rounded-md border border-zinc-300 px-2 py-1"
          />
        </label>
        <button
          type="submit"
          className="rounded-md border border-dashed border-zinc-300 px-3 py-1 text-zinc-600 hover:bg-zinc-50"
        >
          この期間で表示
        </button>
      </form>
      <p className="text-xs text-zinc-400">
        未指定時はログイン時点の年の1/1以降で最初に来る日曜日を起点に53週間分を表示
      </p>

      {error && (
        <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      )}

      <div className="overflow-x-auto pb-1">
        <div
          className="grid w-max grid-flow-col gap-[3px]"
          style={{ gridTemplateRows: "repeat(7, 14px)" }}
          aria-busy={loading}
        >
          {Array.from({ length: weekCount * 7 }).map((_, i) => {
            const date = dates[i];
            const hours = date ? hoursByDate[date] ?? 0 : undefined;
            if (!date) return <div key={i} className="h-[14px] w-[14px]" />;
            return (
              <div
                key={date}
                title={`${date}: ${hours}h`}
                className={`h-[14px] w-[14px] rounded-sm border border-zinc-200 ${LEVEL_CLASS[levelFor(hours ?? 0)]}`}
              />
            );
          })}
        </div>
      </div>

      <div className="flex items-center gap-1 text-xs text-zinc-400">
        <span>少ない</span>
        {[0, 1, 2, 3, 4].map((lv) => (
          <span key={lv} className={`h-3 w-3 rounded-sm border border-zinc-200 ${LEVEL_CLASS[lv]}`} />
        ))}
        <span>多い(その日の実績工数合計・時間)</span>
      </div>
    </section>
  );
}

// デフォルトの集計期間(当日が属する年の1/1以降で最初に来る日曜日から53週間/371日)を計算する。
// バックエンド側の既定ロジック(internal/handlers/tasks.goのparseCalendarRange、ADR-0010)と
// 同じ規則をフロントエンドでも再現し、グリッドの列数とAPIへ渡すfrom/toを常に一致させる。

function toDateString(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export function defaultCalendarRange(today = new Date()): { from: string; to: string } {
  const year = today.getUTCFullYear();
  const jan1 = new Date(Date.UTC(year, 0, 1));
  const offsetToSunday = (7 - jan1.getUTCDay()) % 7;

  const from = new Date(jan1);
  from.setUTCDate(jan1.getUTCDate() + offsetToSunday);

  const to = new Date(from);
  to.setUTCDate(from.getUTCDate() + 370);

  return { from: toDateString(from), to: toDateString(to) };
}

export function enumerateDates(from: string, to: string): string[] {
  const dates: string[] = [];
  const cur = new Date(`${from}T00:00:00Z`);
  const end = new Date(`${to}T00:00:00Z`);
  while (cur <= end) {
    dates.push(toDateString(cur));
    cur.setUTCDate(cur.getUTCDate() + 1);
  }
  return dates;
}

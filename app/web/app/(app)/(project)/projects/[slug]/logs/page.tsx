export default async function LogsPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return (
    <div className="p-8">
      <h1 className="text-text-primary text-2xl font-semibold">Logs</h1>
      <p className="text-text-muted mt-1 font-mono text-sm">{slug}</p>
    </div>
  );
}

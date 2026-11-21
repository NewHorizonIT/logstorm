import { LogsContent } from "@/components/logs/logs-content";

export default async function LogsPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <LogsContent slug={slug} />;
}

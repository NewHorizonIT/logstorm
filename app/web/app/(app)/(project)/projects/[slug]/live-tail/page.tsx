import { LiveTailContent } from "@/components/live-tail/live-tail-content";

export default async function LiveTailPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <LiveTailContent slug={slug} />;
}

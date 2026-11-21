import { ApiKeysContent } from "@/components/api-keys/api-keys-content";

export default async function ApiKeysPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <ApiKeysContent slug={slug} />;
}

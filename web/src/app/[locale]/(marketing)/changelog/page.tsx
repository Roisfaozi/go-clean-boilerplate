import { type Change, changes } from "content";
import dayjs from "dayjs";
import { type Metadata } from "next";

function ChangeCard(change: Change) {
  return (
    <article className="prose prose-slate dark:prose-invert mb-8">
      <h2 className="mb-0 text-3xl font-semibold tracking-tight transition-colors">
        {change.title}
      </h2>
      <time className="text-muted-foreground text-sm" dateTime={change.date}>
        {dayjs(change.date).format("MMM DD YYYY")}
      </time>
      <div dangerouslySetInnerHTML={{ __html: change.content }} />
    </article>
  );
}

export const metadata: Metadata = {
  title: "Changelog",
  description: "All the latest updates, improvements, and fixes.",
};

export default function Changelog() {
  const posts = changes.sort((a, b) =>
    dayjs(a.date).isAfter(dayjs(b.date)) ? -1 : 1
  );

  return (
    <div className="container min-h-screen py-8">
      <h1 className="text-4xl font-bold tracking-tight lg:text-5xl">
        Changelog
      </h1>
      <p className="text-muted-foreground mt-2.5 mb-10 text-xl">
        All the latest updates, improvements, and fixes.
      </p>
      <div className="space-y-10">
        {posts.map((change, idx) => (
          <ChangeCard key={idx} {...change} />
        ))}
      </div>
    </div>
  );
}

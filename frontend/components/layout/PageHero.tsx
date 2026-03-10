type PageHeroMetric = {
  label: string;
  value: string;
};

export default function PageHero({
  eyebrow,
  title,
  description,
  metrics,
}: {
  eyebrow: string;
  title: string;
  description: string;
  metrics: PageHeroMetric[];
}) {
  return (
    <section className="page-hero">
      <div>
        <p className="page-hero__eyebrow">{eyebrow}</p>
        <h1 className="page-hero__title">{title}</h1>
        <p className="page-hero__description">{description}</p>
      </div>

      <div className="page-hero__metrics">
        {metrics.map((metric) => (
          <div key={metric.label} className="page-hero__metric">
            <span>{metric.label}</span>
            <strong>{metric.value}</strong>
          </div>
        ))}
      </div>
    </section>
  );
}

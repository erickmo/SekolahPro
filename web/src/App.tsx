import {
  ArrowRight,
  BadgeCheck,
  BarChart3,
  BookOpen,
  Building2,
  Check,
  CreditCard,
  MessagesSquare,
  ShieldCheck,
  Sparkles,
  Users,
} from 'lucide-react'

const metrics = [
  { label: 'Schools onboarded', value: '120+' },
  { label: 'Admin work reduced', value: '45%' },
  { label: 'Parent response rate', value: '3x' },
]

const features = [
  {
    icon: Building2,
    title: 'Unified school operations',
    description: 'Manage admissions, classes, staff, schedules, and campus workflows from one dashboard.',
  },
  {
    icon: CreditCard,
    title: 'Billing and finance automation',
    description: 'Track invoices, tuition payments, reminders, and financial reports without spreadsheet chaos.',
  },
  {
    icon: MessagesSquare,
    title: 'Parent communication hub',
    description: 'Send announcements, attendance updates, and academic progress through a single channel.',
  },
  {
    icon: BarChart3,
    title: 'Real-time academic insight',
    description: 'Monitor grades, attendance, teacher performance, and school growth with live analytics.',
  },
]

const benefits = [
  'Fast onboarding for admin and teachers',
  'Secure role-based access for every team',
  'Works for growing private and public schools',
  'Clear reports for principals and founders',
]

const testimonials = [
  {
    quote: 'SekolahPro helped us replace disconnected tools with one workflow that our staff actually enjoys using.',
    name: 'Nadia Putri',
    role: 'School Director',
  },
  {
    quote: 'Billing follow-up became dramatically easier, and parents now receive updates much faster than before.',
    name: 'Rizky Saputra',
    role: 'Finance Manager',
  },
]

export function App() {
  return (
    <div className="pageShell">
      <header className="topbar">
        <div className="brand">
          <div className="brandMark">S</div>
          <span>SekolahPro</span>
        </div>
        <nav className="navLinks">
          <a href="#features">Features</a>
          <a href="#benefits">Benefits</a>
          <a href="#testimonials">Stories</a>
        </nav>
        <div className="navActions">
          <a className="button buttonGhost" href="#features">
            Explore
          </a>
          <a className="button buttonPrimary" href="#contact">
            Book demo
          </a>
        </div>
      </header>

      <main>
        <section className="heroSection">
          <div className="heroCopy">
            <div className="pill">
              <Sparkles size={14} />
              Built for modern schools
            </div>
            <h1>
              Run your school with one elegant operating system.
            </h1>
            <p>
              SekolahPro brings academics, finance, communication, and operations into one SaaS platform so your team can move faster and serve students better.
            </p>
            <div className="heroActions">
              <a className="button buttonPrimary" href="#contact">
                Start free consultation
                <ArrowRight size={16} />
              </a>
              <a className="button buttonSecondary" href="#features">
                See platform highlights
              </a>
            </div>
            <div className="metricGrid">
              {metrics.map((metric) => (
                <div className="metricCard" key={metric.label}>
                  <strong>{metric.value}</strong>
                  <span>{metric.label}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="heroPanel">
            <div className="dashboardCard">
              <div className="dashboardHeader">
                <div>
                  <span className="eyebrow">Live school overview</span>
                  <h2>Everything leadership needs at a glance</h2>
                </div>
                <BadgeCheck size={22} />
              </div>
              <div className="dashboardStats">
                <div>
                  <span>Attendance today</span>
                  <strong>96.4%</strong>
                </div>
                <div>
                  <span>Collected this month</span>
                  <strong>Rp248M</strong>
                </div>
                <div>
                  <span>Unread parent chats</span>
                  <strong>18</strong>
                </div>
              </div>
              <div className="dashboardList">
                <div>
                  <Users size={18} />
                  <span>Admissions pipeline updates automatically</span>
                </div>
                <div>
                  <BookOpen size={18} />
                  <span>Class schedules and grading stay in sync</span>
                </div>
                <div>
                  <ShieldCheck size={18} />
                  <span>Permissions keep finance and academic data protected</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="section" id="features">
          <div className="sectionHeading">
            <span>Platform features</span>
            <h2>One platform, every critical school workflow</h2>
          </div>
          <div className="featureGrid">
            {features.map((feature) => {
              const Icon = feature.icon
              return (
                <article className="featureCard" key={feature.title}>
                  <div className="featureIcon">
                    <Icon size={20} />
                  </div>
                  <h3>{feature.title}</h3>
                  <p>{feature.description}</p>
                </article>
              )
            })}
          </div>
        </section>

        <section className="section benefitSection" id="benefits">
          <div className="benefitPanel">
            <span>Why schools choose SekolahPro</span>
            <h2>Reduce admin burden and create a better parent experience.</h2>
            <div className="benefitList">
              {benefits.map((benefit) => (
                <div className="benefitItem" key={benefit}>
                  <Check size={16} />
                  <span>{benefit}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="quotePanel">
            <div className="quoteCard">
              <span>Implementation in weeks, not months</span>
              <strong>Built to scale from one campus to a growing education group.</strong>
            </div>
            <div className="quoteCard altCard">
              <span>Clear ROI</span>
              <strong>Improve payment collection, staff efficiency, and parent trust in one move.</strong>
            </div>
          </div>
        </section>

        <section className="section" id="testimonials">
          <div className="sectionHeading">
            <span>Customer stories</span>
            <h2>Trusted by teams that want better systems</h2>
          </div>
          <div className="testimonialGrid">
            {testimonials.map((testimonial) => (
              <article className="testimonialCard" key={testimonial.name}>
                <p>“{testimonial.quote}”</p>
                <div>
                  <strong>{testimonial.name}</strong>
                  <span>{testimonial.role}</span>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className="section ctaSection" id="contact">
          <div>
            <span>Ready to grow with SekolahPro?</span>
            <h2>Launch a school experience that feels modern, efficient, and connected.</h2>
          </div>
          <a className="button buttonPrimary" href="mailto:hello@sekolahpro.com">
            Contact sales
            <ArrowRight size={16} />
          </a>
        </section>
      </main>
    </div>
  )
}

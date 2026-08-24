<script setup>
import IconWindows from '@/icons/IconWindows.vue';
import IconMacos from '@/icons/IconMacos.vue';
import IconLinux from '@/icons/IconLinux.vue';
import { PhArrowUpRight, PhUserCircle, PhStar } from '@phosphor-icons/vue';

const onBtnHover = (event) => {
  const btn = event.currentTarget;
  btn.style.transform = 'translateY(-4px) scale(1.02)';
};

const onBtnLeave = (event) => {
  const btn = event.currentTarget;
  btn.style.transform = 'translateY(0) scale(1)';
};
</script>

<template>
  <section id="download" class="download-section">
    <div class="download-background">
      <div class="gradient-layer layer-1"></div>
      <div class="gradient-layer layer-2"></div>
    </div>

    <div class="container">
      <div class="download-content">
        <div class="content-text">
          <h2 class="download-title">
            <span class="title-line">Today, Let's</span>
            <span class="title-line gradient-text">Reading Smarter</span>
          </h2>
          <p class="download-description">
            Join thousands of users who have already transformed their RSS reading experience.
            Download MrRSS and harness the power of AI.
          </p>

          <div class="download-buttons">
            <a
              href="https://github.com/DevXDojo/MrRSS/releases"
              class="download-btn windows"
              @mouseenter="onBtnHover"
              @mouseleave="onBtnLeave"
            >
              <div class="btn-icon">
                <IconWindows class="w-8 h-8" />
              </div>
              <div class="btn-content">
                <span class="btn-label">Download for</span>
                <span class="btn-platform">Windows</span>
              </div>
              <PhArrowUpRight class="btn-arrow" weight="bold" />
            </a>

            <a
              href="https://github.com/DevXDojo/MrRSS/releases"
              class="download-btn macos"
              @mouseenter="onBtnHover"
              @mouseleave="onBtnLeave"
            >
              <div class="btn-icon">
                <IconMacos class="w-8 h-8" />
              </div>
              <div class="btn-content">
                <span class="btn-label">Download for</span>
                <span class="btn-platform">macOS</span>
              </div>
              <PhArrowUpRight class="btn-arrow" weight="bold" />
            </a>

            <a
              href="https://github.com/DevXDojo/MrRSS/releases"
              class="download-btn linux"
              @mouseenter="onBtnHover"
              @mouseleave="onBtnLeave"
            >
              <div class="btn-icon">
                <IconLinux class="w-8 h-8" />
              </div>
              <div class="btn-content">
                <span class="btn-label">Download for</span>
                <span class="btn-platform">Linux</span>
              </div>
              <PhArrowUpRight class="btn-arrow" weight="bold" />
            </a>
          </div>
        </div>
      </div>

      <!-- Testimonials Section -->
      <div class="testimonials-section">
        <h3 class="testimonials-title">What Users Say</h3>
        <div class="testimonials-grid">
          <div class="testimonial-card">
            <div class="testimonial-header">
              <div class="testimonial-avatar">
                <PhUserCircle weight="fill" size="64" />
              </div>
              <div class="testimonial-info">
                <span class="testimonial-name">Sarah Chen</span>
                <span class="testimonial-role">Tech Journalist</span>
              </div>
            </div>
            <p class="testimonial-text">
              "MrRSS has completely transformed how I consume news. The AI summaries are incredibly
              accurate and save me hours every week. The local processing gives me peace of mind
              about my data."
            </p>
            <div class="testimonial-rating">
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
            </div>
          </div>

          <div class="testimonial-card">
            <div class="testimonial-header">
              <div class="testimonial-avatar">
                <PhUserCircle weight="fill" size="64" />
              </div>
              <div class="testimonial-info">
                <span class="testimonial-name">Marcus Rodriguez</span>
                <span class="testimonial-role">Software Developer</span>
              </div>
            </div>
            <p class="testimonial-text">
              "As a developer, I appreciate the open-source nature and the fact that everything runs
              locally. The translation feature helps me keep up with global tech news effortlessly."
            </p>
            <div class="testimonial-rating">
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
            </div>
          </div>

          <div class="testimonial-card">
            <div class="testimonial-header">
              <div class="testimonial-avatar">
                <PhUserCircle weight="fill" size="64" />
              </div>
              <div class="testimonial-info">
                <span class="testimonial-name">Emily Watson</span>
                <span class="testimonial-role">Research Analyst</span>
              </div>
            </div>
            <p class="testimonial-text">
              "The smart discovery feature is phenomenal. It's like having a personal research
              assistant that knows exactly what I'm interested in. Highly recommended for anyone who
              reads a lot!"
            </p>
            <div class="testimonial-rating">
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
              <PhStar weight="fill" :size="16" color="#FFC107" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.download-section {
  position: relative;
  padding: 8rem 2rem;
  background: linear-gradient(180deg, #f8f9fa 0%, #ffffff 100%);
  overflow: hidden;
}

.download-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.gradient-layer {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.12;
}

.layer-1 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, #007bff 0%, #0056b3 100%);
  top: -200px;
  left: -100px;
}

.layer-2 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, #ffc107 0%, #ff9800 100%);
  bottom: -150px;
  right: -100px;
}

.container {
  position: relative;
  max-width: 1200px;
  margin: 0 auto;
  z-index: 1;
}

.download-content {
  text-align: center;
  margin-bottom: 6rem;
}

.content-text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
}

.download-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: clamp(2.5rem, 5vw, 4rem);
  font-weight: 700;
  line-height: 1.1;
  text-align: center;
}

.title-line {
  display: block;
  color: #1a1a1a;
}

.gradient-text {
  background: linear-gradient(135deg, #007bff 0%, #ffc107 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  background-size: 200% 200%;
  animation: gradientShift 3s ease-in-out infinite;
}

@keyframes gradientShift {
  0%,
  100% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
}

.download-description {
  font-size: 1.125rem;
  color: rgba(0, 0, 0, 0.65);
  line-height: 1.7;
  max-width: 600px;
}

.download-buttons {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 450px;
}

.download-btn {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1.25rem 2rem;
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  text-decoration: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.download-btn:hover {
  border-color: rgba(0, 123, 255, 0.3);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.download-btn .btn-icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  color: #444444;
  transition: color 0.3s;
}

.btn-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
  text-align: left;
}

.btn-label {
  font-size: 0.875rem;
  color: rgba(0, 0, 0, 0.5);
}

.btn-platform {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1a1a1a;
}

.btn-arrow {
  flex-shrink: 0;
  color: rgba(0, 0, 0, 0.3);
  transition: all 0.3s;
}

.download-btn:hover .btn-arrow {
  color: #007bff;
  transform: translate(4px, -4px);
}

.download-btn:hover .btn-icon {
  color: #007bff;
}

/* Testimonials Section */
.testimonials-section {
  margin-bottom: 6rem;
}

.testimonials-title {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 2rem;
  font-weight: 700;
  color: #1a1a1a;
  text-align: center;
  margin-bottom: 3rem;
}

.testimonials-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.testimonial-card {
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 20px;
  padding: 2rem;
  transition: all 0.3s ease;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.testimonial-card:hover {
  border-color: rgba(0, 123, 255, 0.2);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.08);
  transform: translateY(-4px);
}

.testimonial-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.testimonial-avatar {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(0, 123, 255, 0.1), rgba(0, 123, 255, 0.05));
  border: 1px solid rgba(0, 123, 255, 0.2);
  border-radius: 50%;
  color: #007bff;
}

.testimonial-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.testimonial-name {
  font-weight: 600;
  color: #1a1a1a;
  font-size: 1rem;
}

.testimonial-role {
  font-size: 0.875rem;
  color: rgba(0, 0, 0, 0.5);
}

.testimonial-text {
  flex: 1;
  color: rgba(0, 0, 0, 0.7);
  line-height: 1.6;
  margin-bottom: 1.5rem;
  font-size: 0.9375rem;
}

.testimonial-rating {
  display: flex;
  gap: 0.25rem;
}

@media (max-width: 1024px) {
  .testimonials-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .download-section {
    padding: 4rem 1rem;
  }

  .download-description {
    font-size: 1rem;
  }

  .download-btn {
    padding: 1rem 1.5rem;
  }

  .btn-platform {
    font-size: 1.125rem;
  }

  .testimonials-title {
    font-size: 1.5rem;
    margin-bottom: 2rem;
  }
}

/* Handle small height screens */
@media (max-height: 700px) {
  .download-section {
    padding: 3rem 2rem;
  }

  .download-content {
    margin-bottom: 4rem;
  }

  .testimonials-section {
    margin-bottom: 4rem;
  }
}

@media (max-width: 480px) {
  .download-section {
    padding: 3rem 0.75rem;
  }

  .download-title {
    font-size: clamp(2rem, 8vw, 3rem);
  }

  .download-description {
    font-size: 0.9375rem;
  }

  .download-buttons {
    max-width: 100%;
  }

  .download-btn {
    padding: 0.875rem 1.25rem;
    gap: 1rem;
  }

  .download-btn .btn-icon {
    width: 48px;
    height: 48px;
  }

  .btn-platform {
    font-size: 1rem;
  }

  .testimonials-title {
    font-size: 1.375rem;
  }

  .testimonial-card {
    padding: 1.5rem;
  }

  .testimonial-header {
    margin-bottom: 1rem;
  }

  .testimonial-avatar {
    width: 40px;
    height: 40px;
  }

  .testimonial-name {
    font-size: 0.9375rem;
  }

  .testimonial-role {
    font-size: 0.8125rem;
  }

  .testimonial-text {
    font-size: 0.875rem;
  }
}
</style>

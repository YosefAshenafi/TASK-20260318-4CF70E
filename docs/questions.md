# questions.md

## 1. Vue Version & Frontend Tooling

**Question:** The prompt says the frontend is built with Vue.js but does not specify Vue version or tooling.
**Assumption:** Use Vue 3 + Vite.
**Solution:** Build frontend using Vue 3 with Vite for modern performance and development speed.

---

## 2. UI Framework / Component Library

**Question:** The prompt describes an admin dashboard UI but does not specify a UI library.
**Assumption:** Use Element Plus.
**Solution:** Implement tables, forms, dialogs, and layouts using Element Plus.

---

## 3. TypeScript Requirement

**Question:** The prompt does not clarify whether the frontend uses JavaScript or TypeScript.
**Assumption:** Use TypeScript.
**Solution:** Use TypeScript for safer DTO typing, better maintainability, and large-scale project structure.

---

## 4. ORM / Database Access Layer

**Question:** The prompt specifies Go + MySQL but does not mention ORM choice.
**Assumption:** Use GORM.
**Solution:** Use GORM for structured entity models, migrations, and cleaner query management.

---

## 5. Authentication Token Format

**Question:** The prompt says "session tokens valid for 8 hours" but does not clarify whether tokens are JWT or stored server-side.
**Assumption:** Use server-side stored session tokens (not JWT).
**Solution:** Generate random tokens, store them in MySQL, validate per request, and invalidate on logout.

---

## 6. PII Encryption Algorithm

**Question:** The prompt requires encrypting sensitive fields (ID number, phone) but does not specify algorithm.
**Assumption:** Use AES-256 encryption.
**Solution:** Encrypt PII fields using AES-256 and store encrypted values in MySQL.

---

## 7. Candidate Match Score Logic

**Question:** The prompt requires a 0–100 match score but does not define the scoring algorithm.
**Assumption:** Score is based mainly on skill match, then experience, then education.
**Solution:** Implement weighted scoring and return explainable breakdown (matched skills, experience met, etc.).

---

## 8. Duplicate Candidate Merge Rules

**Question:** The prompt says duplicates should be merged but does not define merge priority rules.
**Assumption:** The most recently updated record is kept as the base record.
**Solution:** Merge missing fields from older record into the newest one, and keep an audit record of the merge.

---

## 9. Auto-Deactivation Scheduling

**Question:** The prompt requires automatic deactivation on expiration but does not clarify whether this is real-time or scheduled.
**Assumption:** Use a scheduled background job.
**Solution:** Run a daily scheduled process to deactivate expired qualifications automatically.

---

## 10. Deployment Method (Docker vs Bare Metal)

**Question:** The prompt mentions offline intranet deployment but does not specify the deployment approach.
**Assumption:** Use Docker Compose.
**Solution:** Provide Docker Compose setup for Gin backend, MySQL, and optional Nginx reverse proxy.

---

## 11. Application Naming (If Not Explicitly Provided)

**Question:** The prompt defines the business domain and capabilities but does not provide an official application name.
**Assumption:** Use **PharmaOps Talent & Compliance Platform** as the working product name.
**Solution:** Apply this name consistently in design and API documentation unless the client provides a preferred official product name later.
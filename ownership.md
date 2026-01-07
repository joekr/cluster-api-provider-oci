# Cluster API Provider OCI - Team Ownership & Responsibilities

## Overview

The **Cluster API Provider for Oracle Cloud Infrastructure (CAPOCI)** is a Kubernetes infrastructure provider that enables teams to provision and manage Kubernetes clusters on Oracle Cloud using Kubernetes-native APIs. This document outlines the responsibilities required to own and maintain this critical infrastructure component.

### What is CAPOCI?

CAPOCI implements the [Cluster API](https://cluster-api.sigs.k8s.io/) specification, allowing users to:
- Create and manage self-hosted Kubernetes clusters on OCI compute infrastructure
- Provision and manage Oracle Kubernetes Engine (OKE) managed clusters
- Automate network infrastructure setup (Virtual Cloud Networks, load balancers, security rules)
- Manage cluster lifecycle operations (creation, scaling, upgrades, deletion) through Kubernetes custom resources

**Technology Stack:** Go, Kubernetes Controllers, OCI SDK, Cluster API

---

## Core Team Responsibilities

### 1. Software Development & Maintenance

#### API Design & Versioning
- **Maintain API Compatibility:** Support multiple API versions (currently v1beta1 legacy and v1beta2 current) with proper conversion webhooks
- **Custom Resource Definitions (CRDs):** Manage 14+ CRD definitions including OCICluster, OCIMachine, OCIManagedCluster, and machine pool resources
- **API Evolution:** Design and implement new API features while maintaining backward compatibility
- **Deprecation Management:** Plan and execute API version deprecations following Kubernetes standards

#### Controller Development
- **Reconciliation Logic:** Develop and maintain Kubernetes controllers that reconcile desired state (defined in custom resources) with actual OCI infrastructure
- **Error Handling:** Implement robust retry logic, exponential backoff, and graceful degradation for OCI API failures
- **State Management:** Ensure controllers properly track infrastructure state and handle drift detection
- **Performance Optimization:** Optimize controller performance for managing large numbers of clusters and machines

#### OCI Service Integration
- **Core Services:** Maintain integrations with essential OCI services:
  - **Compute:** Instance creation, configuration, and lifecycle management
  - **Networking (VCN):** Virtual networks, subnets, route tables, internet/NAT gateways, security lists, network security groups
  - **Load Balancing:** Both classic load balancers and network load balancers for control plane and workload access
  - **Container Engine (OKE):** Managed Kubernetes cluster provisioning and management
  - **Identity (IAM):** Authentication, authorization, and tenancy management
- **New Service Integration:** Evaluate and implement new OCI services as they become relevant to Kubernetes infrastructure
- **API Updates:** Track OCI SDK releases and update integrations when APIs change

#### Code Quality & Testing
- **Unit Testing:** Maintain comprehensive unit test coverage across all packages using Ginkgo/Gomega framework
- **Mock Testing:** Create and maintain mocks for all OCI service interactions using gomock
- **Code Reviews:** Review all code changes for correctness, performance, security, and maintainability
- **Linting:** Ensure code passes golangci-lint checks and adheres to Go best practices
- **Refactoring:** Continuously improve code organization and eliminate technical debt

---

### 2. Testing & Quality Assurance

#### End-to-End (E2E) Testing
- **Test Coverage:** Maintain E2E tests for all major cluster configurations:
  - Self-managed clusters with various CNI plugins (Calico, Antrea, Flannel)
  - OKE managed clusters with self-managed and virtual node pools
  - Multi-region clusters and VCN peering scenarios
  - Windows workload support
  - IPv6 networking configurations
  - Machine pools and instance pools
- **Test Infrastructure:** Maintain test OCI tenancies, compartments, and test data
- **Test Automation:** Ensure E2E tests run reliably in CI pipeline
- **Debugging:** Investigate and resolve test failures, including flaky tests

#### Kubernetes Conformance Testing
- **CNCF Conformance:** Run and validate Kubernetes conformance test suites against CAPOCI-provisioned clusters
- **Version Testing:** Test against all supported Kubernetes versions (currently v1.29-v1.31+)
- **Certification:** Maintain any required Kubernetes conformance certifications

#### Integration Testing
- **Cluster API Integration:** Validate compatibility with upstream Cluster API releases
- **Add-on Testing:** Test integration with common Kubernetes add-ons (CSI drivers, cloud controller managers, ingress controllers)
- **Upgrade Testing:** Validate cluster upgrade paths for both Kubernetes versions and CAPOCI versions

#### Performance & Scale Testing
- **Cluster Scale:** Test provisioning and managing clusters with large node counts
- **Multi-Cluster:** Validate managing multiple clusters from a single management cluster
- **Resource Limits:** Identify and document scaling limits and bottlenecks
- **Performance Benchmarks:** Establish and track performance metrics (cluster creation time, reconciliation latency)

---

### 3. Release Management

#### Version Planning
- **Semantic Versioning:** Follow semantic versioning (MAJOR.MINOR.PATCH) for all releases
- **Release Cadence:** Establish and maintain regular release schedule aligned with Cluster API upstream releases
- **Release Notes:** Document all changes, new features, bug fixes, and breaking changes
- **Compatibility Matrix:** Maintain version compatibility documentation for Kubernetes, Cluster API, and OCI services

#### Build & Publish Process
- **Multi-Architecture Builds:** Build and publish Docker images for AMD64 and ARM64 architectures
- **Container Registry:** Publish official images to ghcr.io/oracle/cluster-api-oci-controller
- **Artifact Generation:** Generate and publish deployment manifests (infrastructure-components.yaml)
- **Template Publishing:** Publish cluster templates for all supported configurations (23+ templates)
- **Metadata Management:** Update metadata.yaml for clusterctl compatibility

#### Release Automation
- **CI/CD Pipeline:** Maintain GitHub Actions workflows for automated releases
- **Image Signing:** Implement and maintain container image signing for security
- **Artifact Verification:** Ensure release artifacts are validated before publication
- **Rollback Procedures:** Document and test rollback procedures for failed releases

#### Hotfix Management
- **Critical Bugs:** Rapidly develop, test, and release hotfixes for critical production issues
- **Security Patches:** Expedite releases for security vulnerabilities
- **Backporting:** Backport critical fixes to supported older versions when necessary

---

### 4. Documentation

#### User Documentation
- **Getting Started Guides:** Maintain clear, step-by-step instructions for new users
- **Installation Documentation:** Document all installation methods (clusterctl, manual deployment)
- **Configuration Reference:** Comprehensive documentation of all API fields and configuration options
- **Networking Guides:** Detailed networking architecture documentation (VCN design, CNI selection, load balancer configuration)
- **Use Case Examples:** Provide documented examples for common scenarios (production clusters, development environments, multi-region setups)

#### Operational Documentation
- **Troubleshooting Guides:** Document common issues and their resolutions
- **Upgrade Guides:** Step-by-step procedures for upgrading CAPOCI and managed clusters
- **Disaster Recovery:** Document backup, restore, and recovery procedures
- **Monitoring & Observability:** Guide users on implementing monitoring and logging
- **Performance Tuning:** Best practices for optimizing cluster performance

#### Developer Documentation
- **Architecture Documentation:** Document system architecture, component interactions, and design decisions
- **Contributing Guide:** Clear instructions for external and internal contributors
- **Development Setup:** Local development environment setup and workflow
- **Testing Guide:** How to run and write tests
- **Code Generation:** Document use of controller-gen, conversion-gen, and other code generation tools

#### API Reference
- **Auto-Generated Docs:** Maintain automated API reference documentation from CRD definitions
- **Field Documentation:** Ensure all API fields have clear descriptions
- **Examples:** Provide working examples for all major API resources
- **Migration Guides:** Document migration paths between API versions

#### Documentation Infrastructure
- **mdBook Maintenance:** Maintain the mdBook-based documentation site
- **GitHub Pages:** Ensure documentation is published correctly to GitHub Pages
- **Documentation CI:** Maintain automation for building and publishing documentation
- **Search Functionality:** Ensure documentation is easily searchable

---

### 5. Infrastructure & Operations

#### OCI Resource Management
- **Network Topology:** Design and implement VCN layouts, subnet allocation, routing, and security group rules
- **Load Balancer Configuration:** Configure and optimize load balancers for control plane HA and workload distribution
- **Compute Resources:** Manage instance shapes, boot volumes, and custom images
- **Identity Management:** Configure proper IAM policies, compartments, and user principals for CAPOCI operations

#### Multi-Tenancy Support
- **Tenant Isolation:** Ensure proper isolation between different tenants using the same CAPOCI installation
- **Resource Quotas:** Implement and document resource quota management
- **Security Boundaries:** Maintain clear security boundaries between tenants

#### High Availability & Resilience
- **Controller HA:** Ensure CAPOCI controller is deployed with high availability (leader election)
- **Failure Domain Support:** Implement and test failure domain spreading for cluster nodes
- **Disaster Recovery:** Design and test disaster recovery scenarios
- **Retry Logic:** Implement robust retry mechanisms for transient OCI API failures

#### Monitoring & Observability
- **Metrics Exposure:** Maintain Prometheus metrics for controller health and operations
- **Logging:** Ensure comprehensive, structured logging for troubleshooting
- **Alerting:** Define and document alerting rules for operational issues
- **Dashboards:** Provide reference Grafana dashboards for monitoring

#### Cost Management
- **Resource Optimization:** Identify and document cost-saving opportunities
- **Lifecycle Management:** Implement proper cleanup of unused resources
- **Cost Tracking:** Enable and document cost allocation and tracking

---

### 6. Dependency Management

#### Upstream Cluster API
- **Version Tracking:** Monitor Cluster API releases and compatibility requirements
- **Contract Compliance:** Ensure compliance with Cluster API provider contracts (currently v1beta1)
- **Upstream Contributions:** Contribute bug fixes and improvements back to Cluster API when appropriate
- **Breaking Changes:** Adapt to breaking changes in upstream Cluster API

#### OCI Go SDK
- **SDK Updates:** Regularly update to latest OCI Go SDK versions (currently v65.81.1)
- **API Changes:** Adapt to OCI service API changes and deprecations
- **Security Updates:** Promptly apply SDK security patches
- **Feature Adoption:** Evaluate and adopt new OCI features exposed through SDK updates

#### Kubernetes Dependencies
- **Version Support:** Maintain compatibility with supported Kubernetes versions (typically last 3-4 minor versions)
- **API Changes:** Track and adapt to Kubernetes API deprecations and changes
- **Client Library Updates:** Keep client-go and related libraries up to date
- **Feature Gates:** Test and support relevant Kubernetes feature gates

#### Third-Party Dependencies
- **CNI Plugins:** Validate compatibility with Calico, Antrea, and Flannel releases
- **CSI Drivers:** Ensure compatibility with OCI block volume and file storage CSI drivers
- **Cloud Controller Manager:** Maintain compatibility with OCI Cloud Controller Manager
- **Security Scanning:** Regularly scan dependencies for vulnerabilities using tools like Dependabot

#### Go Language Updates
- **Go Version:** Keep Go toolchain current (currently Go 1.23+)
- **Module Management:** Maintain clean go.mod and go.sum files
- **Deprecation Handling:** Address deprecated Go features and libraries

---

### 7. Community & Support

#### Issue Management
- **Triage:** Regularly triage incoming GitHub issues, assign labels and priorities
- **Bug Investigation:** Investigate reported bugs, reproduce issues, and identify root causes
- **Feature Requests:** Evaluate feature requests for feasibility and alignment with project goals
- **Issue Resolution:** Fix bugs and implement approved features in a timely manner
- **Duplicate Detection:** Identify and close duplicate issues

#### Pull Request Management
- **Review Process:** Review all incoming pull requests for code quality, tests, and documentation
- **Community PRs:** Provide helpful feedback to external contributors
- **Approval Requirements:** Ensure PRs meet approval requirements (2-3 maintainer approvals)
- **Merge Strategy:** Maintain clean Git history through appropriate merge strategies
- **Backport Management:** Manage backporting of fixes to release branches

#### Community Engagement
- **Slack Channel:** Monitor and respond to questions in #cluster-api-oci on Kubernetes Slack
- **Office Hours:** Host regular community meetings and office hours
- **Meeting Notes:** Maintain meeting notes and action items
- **Community Growth:** Encourage and mentor new contributors

#### Contributor Experience
- **Onboarding:** Provide clear onboarding documentation for new contributors
- **Mentorship:** Mentor community members interested in becoming maintainers
- **Recognition:** Recognize and celebrate community contributions
- **Contributor Agreement:** Manage Oracle Contributor Agreement (OCA) process

#### User Support
- **Question Answering:** Respond to user questions about installation, configuration, and troubleshooting
- **Best Practices:** Share best practices and architectural guidance
- **Escalation:** Escalate critical user issues to appropriate teams
- **Feedback Collection:** Gather user feedback to inform roadmap decisions

---

### 8. Security

#### Vulnerability Management
- **Disclosure Process:** Monitor and respond to security vulnerability reports sent to secalert_us@oracle.com
- **Vulnerability Assessment:** Assess severity and impact of reported vulnerabilities
- **Patch Development:** Develop and test security patches
- **Coordinated Disclosure:** Coordinate disclosure timing with reporters and Oracle security teams
- **Security Advisories:** Publish security advisories for confirmed vulnerabilities

#### Security Scanning
- **Dependency Scanning:** Regularly scan dependencies for known vulnerabilities
- **Container Scanning:** Scan published container images for vulnerabilities
- **SAST/DAST:** Run static and dynamic application security testing
- **License Compliance:** Ensure all dependencies comply with license requirements

#### Secure Development
- **Code Review:** Review code for security issues (injection vulnerabilities, authentication flaws, etc.)
- **Secrets Management:** Ensure no secrets are committed to repository
- **Webhook Security:** Maintain secure admission webhook configurations
- **RBAC:** Implement least-privilege RBAC for controller and user access

#### Compliance
- **Oracle Security Policies:** Ensure compliance with Oracle security policies and procedures
- **Audit Logging:** Implement and maintain audit logging for security-relevant events
- **Penetration Testing:** Coordinate with security teams for periodic penetration testing
- **Security Training:** Stay current on security best practices and threats

---

### 9. CI/CD Pipeline

#### GitHub Actions Maintenance
- **Workflow Definitions:** Maintain GitHub Actions workflows for builds, tests, and releases
- **Workflow Optimization:** Optimize workflow execution time and resource usage
- **Secret Management:** Manage GitHub Actions secrets securely
- **Runner Management:** Ensure appropriate runners are available and functioning

#### Build Pipeline
- **Docker Builds:** Maintain multi-stage Docker builds for efficiency
- **Multi-Architecture:** Build and test on both AMD64 and ARM64 architectures
- **Build Caching:** Optimize build times through effective caching strategies
- **Build Validation:** Ensure builds are reproducible and validated

#### Test Automation
- **Unit Test Execution:** Run unit tests on every pull request
- **E2E Test Execution:** Run E2E tests on relevant changes and scheduled intervals
- **Test Parallelization:** Optimize test execution through parallelization (currently 3 Ginkgo nodes)
- **Test Environment:** Maintain test infrastructure (OCI tenancies, test clusters)
- **Test Data Management:** Manage test data, credentials, and configurations securely

#### Artifact Management
- **Container Registry:** Manage container images in GitHub Container Registry
- **Image Cleanup:** Implement retention policies for old images
- **Release Artifacts:** Generate and publish all required release artifacts
- **Provenance:** Implement build provenance and attestation

#### Pipeline Monitoring
- **Failure Detection:** Monitor CI/CD pipeline for failures
- **Failure Investigation:** Investigate and resolve pipeline failures
- **Performance Tracking:** Track and optimize pipeline performance metrics
- **Cost Management:** Monitor and optimize CI/CD infrastructure costs

---

### 10. Feature Development & Roadmap

#### Feature Planning
- **Roadmap Development:** Develop and maintain public roadmap aligned with user needs and Cluster API direction
- **Prioritization:** Prioritize features based on user impact, effort, and strategic value
- **Stakeholder Alignment:** Align roadmap with Oracle product strategy and customer requirements
- **Community Input:** Incorporate community feedback into roadmap decisions

#### Current Feature Areas
- **Machine Pools:** Continue development of experimental machine pool support (OCIMachinePool, OCIManagedMachinePool)
- **OKE Integration:** Enhance OKE managed cluster features (virtual nodes, addon management)
- **Networking:** IPv6 support, advanced VCN configurations, VCN peering
- **Windows Support:** Maintain and improve Windows workload support with Calico
- **Multi-Region:** Enhance multi-region and disaster recovery capabilities
- **Custom Networking:** Support for custom VCN configurations and bring-your-own-network scenarios

#### Experimental Features
- **Feature Gates:** Manage feature gates for experimental functionality
- **Graduation Process:** Graduate experimental features to stable APIs following Cluster API conventions
- **Deprecation:** Deprecate and remove unsuccessful experimental features

#### Innovation
- **Technology Evaluation:** Evaluate new technologies and approaches (GitOps, policy engines, service mesh integration)
- **Proof of Concepts:** Develop POCs for significant new features
- **Research:** Stay current with Kubernetes ecosystem trends and innovations

---

### 11. Operational Excellence

#### Observability
- **Metrics:** Expose comprehensive Prometheus metrics for all controller operations
- **Logging:** Structured logging with appropriate verbosity levels
- **Tracing:** Consider distributed tracing for complex reconciliation flows
- **Health Checks:** Implement robust readiness and liveness probes

#### Performance
- **Benchmarking:** Establish performance benchmarks for critical operations
- **Profiling:** Profile controller performance to identify bottlenecks
- **Optimization:** Continuously optimize reconciliation loops and OCI API calls
- **Rate Limiting:** Implement and tune rate limiting to avoid OCI API throttling

#### Reliability
- **Error Handling:** Comprehensive error handling with meaningful error messages
- **Idempotency:** Ensure all operations are idempotent
- **Graceful Degradation:** Handle partial failures gracefully
- **Leader Election:** Maintain robust leader election for controller HA

#### Configuration Management
- **Default Values:** Maintain sensible defaults for all configuration options
- **Validation:** Comprehensive validation of user-provided configurations
- **Migration:** Smooth migration paths for configuration changes
- **Backward Compatibility:** Maintain backward compatibility for configuration options

---

### 12. Cluster API Ecosystem

#### Upstream Engagement
- **SIG Participation:** Participate in Cluster API SIG meetings and discussions
- **Issue Tracking:** Monitor upstream Cluster API issues relevant to CAPOCI
- **Collaboration:** Collaborate with other provider maintainers
- **Proposals:** Submit enhancement proposals for cross-provider concerns

#### Provider Contract Compliance
- **Contract Adherence:** Ensure full compliance with Cluster API provider contract specifications
- **Contract Testing:** Run provider contract tests
- **API Compatibility:** Maintain compatibility with Cluster API's expectations

#### Integration & Compatibility
- **clusterctl:** Ensure seamless integration with clusterctl CLI tool
- **Bootstrap Providers:** Test compatibility with kubeadm and other bootstrap providers
- **Control Plane Providers:** Support integration with various control plane providers
- **Add-on Management:** Consider integration with Cluster API add-on management proposals

---

## Skills & Expertise Required

To effectively own this project, the team should have expertise in:

### Technical Skills
- **Go Programming:** Advanced Go development skills including concurrency, testing, and tooling
- **Kubernetes:** Deep understanding of Kubernetes architecture, APIs, controllers, and operators
- **Cluster API:** Familiarity with Cluster API concepts, architecture, and provider development
- **Oracle Cloud Infrastructure:** Understanding of OCI services (Compute, Networking, OKE, IAM, Load Balancers)
- **Networking:** Knowledge of cloud networking concepts (VCNs, subnets, routing, security groups, load balancing)
- **Container Technologies:** Docker, container registries, image building and optimization
- **CI/CD:** GitHub Actions, automated testing, release automation
- **Infrastructure as Code:** Understanding of declarative infrastructure management

### Operational Skills
- **Debugging:** Strong debugging skills for distributed systems
- **Monitoring:** Experience with Prometheus, Grafana, and observability tools
- **Incident Response:** Ability to respond to and resolve production incidents
- **Performance Tuning:** Experience optimizing distributed systems

### Soft Skills
- **Communication:** Clear written and verbal communication for documentation and community interaction
- **Collaboration:** Ability to work with diverse stakeholders (users, contributors, Oracle teams)
- **Project Management:** Planning, prioritization, and execution of complex projects
- **Mentorship:** Ability to mentor contributors and grow the community

---

## Success Metrics

Teams owning this project should track:

### Quality Metrics
- **Test Coverage:** Maintain >80% unit test coverage
- **E2E Test Success Rate:** >95% E2E test pass rate
- **Bug Backlog:** Keep critical bug backlog below defined threshold
- **Code Review Time:** Median PR review time <48 hours

### Performance Metrics
- **Cluster Creation Time:** Track and optimize time to create clusters
- **Reconciliation Latency:** Monitor controller reconciliation performance
- **API Call Efficiency:** Track and minimize unnecessary OCI API calls

### Community Metrics
- **Issue Response Time:** Respond to issues within 48 hours
- **PR Merge Time:** Merge approved PRs within 5 business days
- **Community Growth:** Track contributors and contributions over time
- **User Satisfaction:** Gather and track user satisfaction through surveys

### Release Metrics
- **Release Cadence:** Meet planned release schedule
- **Release Quality:** Track post-release hotfixes as indicator of release quality
- **Deprecation Notice:** Provide >2 release cycles notice for deprecations

---

## Resources & References

### Official Documentation
- **Project Documentation:** https://oracle.github.io/cluster-api-provider-oci/
- **Repository:** https://github.com/oracle/cluster-api-provider-oci
- **Cluster API Documentation:** https://cluster-api.sigs.k8s.io/

### Community
- **Slack:** #cluster-api-oci on Kubernetes Slack (kubernetes.slack.com)
- **Meetings:** Regular community meetings (see repository for details)

### Development
- **Contributing Guide:** See CONTRIBUTING.md in repository
- **Local Development:** See local-dev.md for setup instructions
- **Testing Guide:** See test/README.md for test execution

### Security
- **Vulnerability Reporting:** secalert_us@oracle.com
- **Security Policy:** See SECURITY.md in repository

---

## Conclusion

Owning the Cluster API Provider for OCI is a significant responsibility that requires a skilled, dedicated team. The project sits at the intersection of Kubernetes, cloud infrastructure, and distributed systems, requiring deep technical expertise across multiple domains.

The team must balance multiple priorities:
- **Innovation:** Developing new features and capabilities
- **Stability:** Maintaining a reliable, production-grade infrastructure provider
- **Community:** Supporting and growing an open-source community
- **Security:** Ensuring the highest security standards
- **Performance:** Delivering excellent performance and user experience

Success requires not just technical excellence, but also strong communication, collaboration, and commitment to open-source principles. The team will be stewards of critical infrastructure that powers Kubernetes deployments for Oracle Cloud users worldwide.

This document serves as a comprehensive guide to the responsibilities involved. Teams should regularly revisit and update this document as the project evolves and new responsibilities emerge.

---

**Document Version:** 1.0
**Last Updated:** 2025-12-17
**Maintained By:** CAPOCI Team

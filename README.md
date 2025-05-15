# RuiQi WAF

<div align="center">
<a href="https://deepwiki.com/HUAHUAI23/RuiQi"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki" width="160" height="30"></a>
  <img src="https://img.shields.io/badge/Go-1.24.1-00ADD8?style=flat&logo=go" alt="Go Version" width="140" height="30">
  <img src="https://img.shields.io/badge/HAProxy-3.0-green?style=flat&logo=haproxy" alt="HAProxy" width="140" height="30">
  <img src="https://img.shields.io/badge/OWASP-Coraza-blue?style=flat" alt="Coraza WAF" width="140" height="30">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat" alt="License" width="140" height="30">
</div>

<br>

A modern web application firewall (WAF) management system built on top of [HAProxy](https://www.haproxy.org/) and [OWASP Coraza WAF](https://github.com/corazawaf/coraza) with the [Coraza SPOA](https://github.com/corazawaf/coraza-spoa) integration. This system provides a comprehensive backend API for managing HAProxy configurations, Coraza WAF rules, and traffic inspection.

## 🌐 Click To Run

run the application in less than 30 seconds,default username: **admin**,default password: **admin123**

[![](https://raw.githubusercontent.com/labring-actions/templates/main/Deploy-on-Sealos.svg)](https://usw.sealos.io/?openapp=system-template%3FtemplateName%3DRuiqi-Waf)

## 📺 Demo Video

https://github.com/user-attachments/assets/f74000d7-d229-4d00-843b-1ba28caeb13d

## 📸 RuiQi WAF Interface Showcase

<div align="center">
  <table cellspacing="0" cellpadding="10" style="border-collapse: separate; border-spacing: 15px; background-color: #f8f9fa;">
    <tr>
      <td align="center" style="border-radius: 8px; border: 1px solid #ddd; padding: 12px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
        <img width="767" alt="RuiQi WAF Dashboard" src="./doc/image/waf-1.png" style="border-radius: 6px; max-width: 100%;">
      </td>
      <td align="center" style="border-radius: 8px; border: 1px solid #ddd; padding: 12px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
        <img width="767" alt="RuiQi WAF Rule Management" src="./doc/image/waf-2.png" style="border-radius: 6px; max-width: 100%;">
      </td>
    </tr>
    <tr>
      <td align="center" style="border-radius: 8px; border: 1px solid #ddd; padding: 12px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
        <img width="767" alt="RuiQi WAF Analytics" src="./doc/image/waf-3.png" style="border-radius: 6px; max-width: 100%;">
      </td>
      <td align="center" style="border-radius: 8px; border: 1px solid #ddd; padding: 12px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
        <img width="767" alt="RuiQi WAF Security Config" src="./doc/image/waf-4.png" style="border-radius: 6px; max-width: 100%;">
      </td>
    </tr>
  </table>
</div>

## Core Architecture

Simple WAF implements a modular architecture with HAProxy at the front handling traffic and multiple security engines providing protection:

1. **Coraza WAF Engine**: OWASP ModSecurity-compatible filtering
2. **MicroEngine**: Rule-based matching engine for IP filtering, URL checking, and complex conditional logic
3. **Geographic Analysis**: Location-based traffic filtering
4. **Rate Limiting**: Traffic control and request throttling

The system uses a plugin architecture that allows for continuous enhancement with new security modules.

```mermaid
graph TD
    Client[Client] -->|HTTP Request| HAProxy
    HAProxy -->|TCP Connection| SPOE[Coraza SPOE Agent]
    SPOE -->|Message Type Recognition| TypeCheck
    TypeCheck -->|coraza-req| ReqHandler[Request Handler]
    TypeCheck -->|coraza-res| ResHandler[Response Handler]
    ReqHandler -->|Get App Name| ReqApp[Find Application]
    ResHandler -->|Get App Name| ResApp[Find Application]
    ReqApp -->|Process Request| ReqProcess[Request Processor]
    ResApp -->|Process Response| ResProcess[Response Processor]
    ReqProcess --> Return[Return Results to HAProxy]
    ResProcess --> Return
    HAProxy -->|Apply Action| Action[Allow/Deny/Log]
    Action -->|Response| Client
```

### SPOE Communication Workflow

```
[HAProxy Request] → [internal.Agent.Serve(Listener)]
                          ↓
                   Create spop.Agent
                   agent := spop.Agent{
                       Handler: a,
                       BaseContext: a.Context,
                   }
                          ↓
                [spop.Agent.Serve(Listener)]
                          ↓
                   Accept new connections
                   nc, err := l.Accept()
                          ↓
                   Create protocol handler
                   p := newProtocolClient(ctx, nc, as, handler)
                          ↓
                   Start goroutine for connection
                   go func() {
                       p.Serve()
                   }()
                          ↓
                [protocolClient.Serve]
                   Process frames in connection
                          ↓
                [frameHandler processes Frame]
                   Dispatch based on frame type
                          ↓
                [onNotify handles messages]
                   Create message scanner and objects
                   Call Handler.HandleSPOE
                          ↓
                [internal.Agent.HandleSPOE processing]
                          ↓
                   Parse message type (coraza-req/coraza-res)
                          ↓
                   Get application name
                          ↓
                   Find Application
                          ↓
                   Execute message handler
                          ↓
                   Process return results
                          ↓
                [Return to HAProxy]
```

## Features

- **Multi-Engine Protection**

  - **Coraza WAF Engine**:

    - OWASP Core Rule Set (CRS) support
    - ModSecurity SecLang rule compatibility
    - Custom rule management

  - **MicroEngine**:

    - Rule-based matching for IP, URL, and request path
    - Complex condition combinations (AND/OR logic)
    - IP blacklist/whitelist with CIDR support
    - Efficient regex matching with caching

  - **Geographic Analysis**:

    - Country and region-based filtering
    - Geographic attack visualization

  - **Traffic Control**:
    - Rate limiting and request throttling
    - Connection control mechanisms

- **HAProxy Integration**

  - Full HAProxy lifecycle management (start, stop, restart)
  - Dynamic configuration generation
  - Real-time status monitoring

- **Advanced Security**

  - HTTP request inspection
  - HTTP response inspection
  - Real-time attack detection and prevention
  - RBAC user permission system

- **Monitoring and Logging**

  - WAF attack logs and analytics
  - Traffic statistics
  - Performance metrics

- **API-Driven Workflow**
  - RESTful API with Gin framework
  - Swagger/ReDoc API documentation
  - JWT authentication

## Prerequisites

- Go 1.24.1 or higher
- Node.js 23.10.0 and pnpm 10.6.5 (for frontend development)
- HAProxy 3.0 (for local development)
- MongoDB 6.0
- Docker and Docker Compose (for containerized deployment)

## Local Development

1. Clone the repository:

```bash
git clone https://github.com/HUAHUAI23/RuiQi.git
cd RuiQi
```

2. Setup the frontend development environment:

```bash
cd server/web
pnpm install
pnpm dev # For development mode with hot reload
# or
pnpm build # For production build
cd ../..
```

3. Configure backend environment:

```bash
cp server/.env.template server/.env
# Edit .env with your configurations
```

4. Run the Go backend service:

```bash
go work use ./coraza-spoa ./pkg ./server
cd server
go run main.go
```

The development server will start with:

- API server: `http://localhost:2333/api/v1`
- Swagger UI: `http://localhost:2333/swagger/index.html`
- ReDoc UI: `http://localhost:2333/redoc`
- Frontend: `http://localhost:2333/`

## Docker Deployment

1. Clone the repository:

```bash
git clone https://github.com/HUAHUAI23/RuiQi.git
cd RuiQi
```

2. Build the Docker image:

```bash
docker build -t ruiqi-waf:latest .
```

3. Run as a standalone container:

```bash
docker run -p 2333:2333 -p 8080:8080 -p 443:443 -p 80:80 -p 9443:9443 -p 8404:8404 ruiqi-waf:latest
```

4. Alternatively, use Docker Compose for a complete deployment with MongoDB:

```bash
# Edit docker-compose.yaml to configure environment variables if needed
docker-compose up -d
```

This will start both MongoDB and Simple WAF services with all required configurations.

## Roadmap

Our upcoming features include:

1. **Monitoring Dashboard**

   - Real-time attack visualization map
   - Geographic attack origin analytics
   - Comprehensive security metrics dashboard

2. **Advanced Traffic Control**

   - Fine-grained rate limiting
   - Request frequency analysis
   - Adaptive throttling based on traffic patterns

3. **Alert Integration**

   - Webhook alert system
   - Integration with popular messaging platforms
   - Customizable alert templates

4. **AI Security Analysis**

   - ML-based attack pattern detection
   - AI-assisted rule generation
   - Automated ModSecurity directive creation
   - MCP (Model Control Plane) integration

5. **Enhanced Rule Management**
   - OWASP Top 10 specific rule templates
   - Rule effectiveness scoring
   - One-click protection profiles

For detailed technical documentation on engine architecture and implementation, see the proposals in the `/doc/proposal` directory.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [OWASP Coraza WAF](https://github.com/corazawaf/coraza)
- [Coraza SPOA](https://github.com/corazawaf/coraza-spoa)
- [HAProxy](https://www.haproxy.org/)
- [Go Gin Framework](https://github.com/gin-gonic/gin)

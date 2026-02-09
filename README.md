## Intro
Welcome to ez2boot. This is a self hosted web application designed to provide a simple interface for your colleagues to start and stop your public cloud servers, on demand. Cloud based servers are billed hourly. This is an expected cost for 24/7 production use cases but what about non-production? Often, non-production servers are used in an ad-hoc manner by those who may not have permissions or knowledge to access the native cloud console and start the required servers as needed. Perhaps this means developers, QA teams, sales reps etc. What if they forget to turn them off afterwards, leading to unexpected cloud costs? This project aims to solve this challenge in a secure, user-friendly and compliant way.

## Features
- Simple setup, intended to run as a docker container within your cloud environment.
- User accounts to enforce authenticated access only.
- RBAC to give administrators control over user capabilities.
- Tag-based server selection, allowing Operations teams full control over server availability, and grouping presentation.
- Time-based server sessions. Users choose for how long they want a server group online, and extend or reduce the sesson on demand.
- Clear UI displays indicating the state of server groups, and each server within each group. Reduced user friction and less support required.
- Transparency. All users can see server state allowing teams to work together without uncertainty about server availability.
- Comprehensive, immutable audit logging, showing who did what, and when.
- Customisable user notifications channels, allowing users to opt into automated notifications about their session states.
- Dual-auth. Session-based UI for interactive use, and basic auth API for programmatic control.
- Operations teams can tweak the app's behaviour through environment variables.

## Motivations and inspirations
Like most ideas, this was born out of necessity. My workplace had a use for such an application, but so too would many others. I made the decison to build this using my own time and resources, to give back to the Open Source community which has provided me with many great solutions over the years. 

If you're reading this, you're right on time to be an early adopter. The project is in a beta phase and I would appreciate comments, suggestions and criticisms - it is far from perfect - however every effort is made to minimise bugs and to deliver a functional product. In terms of feature ideas and inspirations, I am drawing from projects such as Prometheus and Uptime Kuma, both of which I use extensively.

## License
This program, ez2boot is licensed under the GNU Affero General Public License v3.0 (AGPLv3).

This license can cause some confusion, but after significant consideration I felt it was necessary to choose this license given the nature of the project. Below are some key points to help dispel confusion. See LICENSE for more details.

- You may use, modify, and distribute this software.
- If you modify this program code or incorporate any part of the program's code and run it as a network service (e.g., in a cloud or docker container, whether internal or publicly accessible), you must make the modified source code of this program available to the users of your service, under AGPLv3.
- **Configuration changes do not count as modifications.**
  For example, editing environment variables, YAML/JSON config files, or toggling options does NOT trigger the obligation to release the source code. 
- **OS or base image updates do not count as modifications.**
  Patching the operating system or libraries (e.g., `apt-get update` inside the container) does NOT trigger the obligation to release the source code.
- This ensures improvements are shared with the community, while routine usage, configuration and maintenance are exempt.
- The ez2boot name, logos, banners, icons, and other branding assets are trademarks of the ez2boot project and are not licensed under the AGPLv3. All rights are reserved.
- Forks created for the purpose of contributing back to this repository may retain the ez2boot name and branding.
- Redistributed or independently published versions must remove or replace them unless prior written permission is obtained.

## Attributions
See THIRD PARTY for attribution and compliance of libraries and components used in this project. If you're the copyright holder of any component used in this project and feel I have not complied with your license, please reach out.
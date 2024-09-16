![Build Status](https://github.com/crystaldb/autodba/actions/workflows/build.yml/badge.svg)

# 🤖 AutoDBA for PostgreSQL 🐘

AutoDBA is an AI agent that runs PostgreSQL databases.
This project exists to ensure that everyone who runs PostgreSQL has access to a skilled virtual database administrator (DBA) at all times.

## 💡 Motivation

We all want our production PostgreSQL databases to run well—they should be reliable, performant, efficient, scalable, and secure.
If your team is fully staffed with database administrators, then you are in luck.
If not, you may find yourself working reactively, dealing with problems as they arise and looking up what to do as you go.

Oftentimes, operational responsibility for databases falls to software engineers and site reliability engineers, who usually have many other things to do.
They have better ways to spend their time than tuning or troubleshooting databases.

Reliability and security are the top priorities for database operations, and our approach reflects that.
The success of autonomous vehicles shows that with careful engineering, one can build systems that operate safely in complex environments.
The time has come to do the same for similarly challenging IT operations—as AutoDBA does for databases.


## 🔍 Project Status: Observability Only

As of today, this project includes only the basic observability features of AutoDBA.
Building an AI agent requires quality data, so we first need to make sure that this foundation is solid.

Why did we put effort into tools for visualizing and exploring data when we envision a future were only machines consume it?
For one, we need to verify that the data is sufficient to support good decision making—if a human expert does not have enough information to make a good decision, then we are probably asking too much of the AI agent if we expect it to do so.
Furthermore, AutoDBA's observability tools add value to the PostgreSQL ecosystem, even without an AI agent.

The core AI agent is under development and will be released once it reaches a reasonable degree of accuracy and stability.


## 🚧 Temporary Limitations

This is an early release of AutoDBA's observability features.
We are committed to supporting PostgreSQL in all environments and popular major versions.
However, the following *temporary limitations* are presently in place:

- Only compatible with PostgreSQL version 16.
- Only works with AWS RDS PostgreSQL.


## 🚀 Installation

### Prerequisites

1. Linux server with network access to your PostgreSQL database
2. A database user with access to read the metrics
3. AWS credentials with permissions to read database metrics

Follow these instructions to install AutoDBA on Linux.

1. Download the latest release of AutoDBA from the [releases page](https://github.com/crystaldb/autodba/releases).

```bash
tar -xzf autodba-<version>-linux-amd64.tar.gz
```

TODO - more to put here


## 🗺️ Roadmap

Our near-term roadmap includes the following:

- Observability
    - [x] Database and system metrics
    - [x] Wait events
    - [ ] Query normalization + PII filtering
    - [ ] Log analysis
    - [ ] Fleet overview
- Alerting
    - [ ] Accurate (low noise) alerting on current status
    - [ ] Predictive health alerts


## 🤝 Contributing

We welcome contributions to AutoDBA! Contributor guidelines are under development so [please reach out](mailto:johann@crystaldb.cloud) if you are interested in working with us.


## 🧑‍💻 About the Authors

AutoDBA is developed by the engineers and database experts at  [CrystalDB](https://www.crystaldb.cloud/).
Our mission is to make it easy for you to run your database well, so you can focus on building better software.
CrystalDB also offers commercial support for AutoDBA and PostgreSQL.


## 📖 Frequently Asked Questions

### What is AutoDBA?

AutoDBA is an AI agent for operating PostgreSQL databases.
This means that it connects to an existing PostgreSQL database and takes actions, as necessary, to ensure that the database remains reliable, efficient, scalable, and secure.


### Will AutoDBA replace my DBA?

Time will tell whether AI agents completely replace human database administrators (DBAs).
Our work suggests that AI agents will do some tasks much better than humans.
They can find patterns across large amounts of data, they are always available, and they respond instantly.
They can also draw upon extensive knowledge bases and operational datasets, allowing them to proceed with less trial and error than people.

On the other hand, people working in the team will have a more nuanced understanding of the needs of the business.
They will be better able to make high-level design decisions and to analyze trade-offs that impact the development process.


### Do I still need to hire a DBA if I use AutoDBA?

If you do not already have a DBA on staff, then chances are good that AutoDBA can allow you to postpone hiring one, particularly if you have platform engineers or site reliability engineers who are interested in applying its recommendations.
CrystalDB and others also offer commercial support for PostgreSQL.


### Which PostgreSQL versions are supported?

Currently, AutoDBA is only compatible with PostgreSQL version 16.
We are working on expanding support to other versions.
Please share your thoughts on how far back we should go.


### Can I use AutoDBA with my on-premises PostgreSQL installation?

At present, AutoDBA only works with AWS RDS PostgreSQL.
Support for on-premises installations, Google Cloud SQL, and Azure SQL is coming.
We want AutoDBA to run anywhere that PostgreSQL runs.


### Will AutoDBA support databases other than PostgreSQL?

At present, we are fully focused on AutoDBA for PostgreSQL.
We expect to maintain that focus for the forseeable future.


### Is AutoDBA open source?

We believe that every PostgreSQL database should be managed by an AI agent.
In pursuit of this vision, we are releasing the core operational features of AutoDBA under the Apache 2.0 open source license.

For avoidance of doubt, AutoDBA is commercial open source software.
Certain enterprise features will be available only in commercial versions of the product.

As of this writing (August 2024), there is active debate what the term “open source” means for AI models.
Is a model open if the developer releases the weights but not the training data and methods?

We are committed to providing open weights and some training data.
However, we also expect to release models trained on proprietary data sets.


### What happens to the data that AutoDBA collects?

AutoDBA collects and stores operational metrics collected from your database.
It does not transmit this data to anyone.
You should keep the web interface secure to avoid exposing this data to others.


### How can I get support for AutoDBA?

CrystalDB offers commercial support for AutoDBA and PostgreSQL.
For more information or to discuss your needs, please contact us at [support@crystaldb.cloud](mailto:support@crystaldb.cloud).


### How can I support the AutoDBA project?

Foremost, use AutoDBA and give us feedback!

We also welcome feature suggestions, bug reports, or contributions to the codebase.

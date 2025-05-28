![CleandGO Banner](docs/assets/top_banner_a.png)

---

**A powerful parser for visual treeviews, converting them into organized file structures.**

---

## **Table of Contents**

1. [About the Project](#about-the-project)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
    - [CLI](#cli)
    - [Examples](#examples)
5. [Roadmap](#roadmap)
6. [Contributing](#contributing)
7. [Contact](#contact)

---

## **About the Project**

CleandGO is an advanced parser for structured visual treeviews. It interprets, saves, and converts these structures into organized files and directories, enabling **automation of directory hierarchy creation**.

**Why CleandGO?**

- 🌟 **Smart Parsing**: Efficiently processes treeview structures.
- 💾 **Serialization and Backup**: Save and recover processed versions.
- 📂 **Automatic File and Directory Creation**: Build hierarchies effortlessly.
- 🔄 **Flexible Input Formats**: Supports various input formats.

---

## **Features**

✅ **Intelligent Treeview Parsing**:

- Processes complex treeview structures.
- Handles diverse input formats.

✅ **Serialization and Backup**:

- Save processed versions for future use.
- Recover treeview structures easily.

✅ **Automated File and Directory Creation**:

- Generates organized hierarchies.
- Ensures consistency and accuracy.

✅ **Flexible Input Support**:

- Compatible with multiple formats.

---

## **Installation**

### **Linux/Mac**

```bash
make install
```

### **Windows**

```powershell
.\install.bat
```

---

## **Usage**

### CLI

Here are some examples of commands you can execute with CleandGO’s CLI:

```bash
# Parse a treeview file
cleandgo parse --source tree_views/tree_test.txt

# Backup a treeview file
cleandgo backup --tree tree_views/tree_test.txt

# Generate files and directories
cleandgo generate --output /my_project

# Serialize a processed version
cleandgo serialize --store version_1.clgo
```

---

## **Examples**

### **1. Parse a Treeview File**

```bash
cleandgo parse --source tree_views/tree_test.txt
```

This processes the treeview and transforms it into valid files and directories.

### **2. Backup a Treeview File**

```bash
cleandgo backup --tree tree_views/tree_test.txt
```

This creates a backup of the treeview for future recovery.

### **3. Generate Files and Directories**

```bash
cleandgo generate --output /my_project
```

This generates organized files and directories based on the treeview structure.

### **4. Serialize a Processed Version**

```bash
cleandgo serialize --store version_1.clgo
```

This saves the processed version for future use.

---

## **Roadmap**

🔜 **Upcoming Features**:

- Support for additional input formats.
- Enhanced treeview visualization.
- Advanced configuration options.

---

## **Contributing**

Contributions are welcome! Feel free to open issues or submit pull requests. Check out the [Contributing Guide](docs/CONTRIBUTING.md) for more details.

---

## **Contact**

💌 **Developer**:

[Byte Sleuth Team](mailto:faelmori@gmail.com)

💼 [Follow me on GitHub](https://github.com/faelmori)

We’re open to new collaborations and feedback. If you find this project interesting, don’t hesitate to reach out!

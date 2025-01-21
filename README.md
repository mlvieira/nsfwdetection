## **NSFW Detector API**

A Go-based service that detects NSFW content in images using a TensorFlow model from [**opennsfw2**](https://github.com/bhky/opennsfw2).

---

## **Setup**

1. **Install Go**

2. **Install TensorFlow C Library:**
Ensure the TensorFlow C library version matches the Python TensorFlow library installed.
```bash
curl -L "https://storage.googleapis.com/tensorflow/versions/2.17.0/libtensorflow-cpu-linux-x86_64.tar.gz" -o libtensorflow.tar.gz
sudo tar -C /usr/local -xzf libtensorflow.tar.gz
sudo ldconfig
```

3. **Set Library Paths:**
```bash
export LIBRARY_PATH=/usr/local/lib
export LD_LIBRARY_PATH=/usr/local/lib
```

4. **Prepare Model:**
Ensure that the TensorFlow Python library version matches the TensorFlow C library version installed earlier.
```bash
cd python
pip install -r requirements.txt
python export_model.py
```

5. **Configure Service:**
```bash
cp config.toml.example config.toml
```
Edit `config.toml` as needed.

6. **Install DB schema**
```bash
mysql -u mysqluser -p database < migrations/schema.sql
```

7. **Build front end**
```bash
npm install && npm run build
```

8. **Configure nginx:**
Copy `nsfw-nginx.conf.example` to `/etc/nginx/sites-enabled/` and edit as needed.

9. **Build and Run:**
```bash
./build.sh
```
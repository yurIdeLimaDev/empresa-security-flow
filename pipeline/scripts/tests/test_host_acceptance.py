from datetime import datetime, timezone, timedelta
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0,str(Path(__file__).parents[1]))
spec=importlib.util.spec_from_file_location("host_acceptance",Path(__file__).parents[1]/"host_acceptance.py")
host=importlib.util.module_from_spec(spec);spec.loader.exec_module(host)


class HostAcceptanceTests(unittest.TestCase):
    def fixture(self,root):
        self.now=datetime.now(timezone.utc); stamp=self.now.isoformat(); self.pin="a"*64
        lock={"tools":[{"name":"tool","execution_class":"automatic","container":{"image":"test/image","digest":"sha256:"+self.pin}},{"name":"restricted","execution_class":"manual"}]}
        values={
            "host-facts.txt":f"checked_at={stamp}\nrunner_sha256={self.pin}\n",
            "host-drift/host-drift.txt":f"checked_at={stamp}\nstatus=passed\ndocker_user_chain=present\ndocker_service=active\nsecurity_updates=enabled\nssh_key_only=passed\ndedicated_roots=passed\nreference_image=present_by_digest\n",
            "supply-chain-report.json":{"checked_at":stamp,"approved":True,"checks":[{"tool":t["name"],"approved":True} for t in lock["tools"]]},
            "runtime-readiness.json":{"checked_at":stamp,"approved":True,"checks":[{"tool":"tool","image":"test/image@sha256:"+self.pin,"status":"passed","observed_output":"sensitive test data"},{"tool":"restricted","status":"restricted"}]},
            "isolation/isolation-smoke.json":{"canary":"negative-blocked-and-positive-allowed","proxy_request":"allowed-pinned-request-logged","proxy_url":"private-value-not-exported"},
        }
        for name,value in values.items():
            path=root/name;path.parent.mkdir(parents=True,exist_ok=True)
            path.write_text(value if isinstance(value,str) else json.dumps(value),encoding="utf-8")
        self.rehash(root)
        backup={"created_at":stamp,"entry_count":2,"ciphertext_sha256":self.pin}
        restore={"restored_at":stamp,"entry_count":2,"backup_sha256":self.pin,"verified":True}
        return lock,backup,restore

    def rehash(self,root):
        rows=[host.digest(p)+"  ./"+p.relative_to(root).as_posix() for p in sorted(root.rglob("*")) if p.is_file() and p.name!="SHA256SUMS"]
        (root/"SHA256SUMS").write_text("\n".join(rows),encoding="utf-8")

    def test_sanitized_pass_and_failures(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);lock,backup,restore=self.fixture(root)
            report=host.evaluate(root,lock,backup,restore,self.pin,now=self.now)
            self.assertEqual(report["status"],"passed")
            self.assertNotIn("private-value",json.dumps(report));self.assertNotIn("sensitive test",json.dumps(report))
            for changed in (dict(restore,verified=False),dict(restore,entry_count=3),dict(restore,backup_sha256="b"*64),dict(restore,restored_at=(self.now-timedelta(days=2)).isoformat())):
                self.assertEqual(host.evaluate(root,lock,backup,changed,self.pin,now=self.now)["status"],"blocked")
            self.assertEqual(host.evaluate(root,lock,backup,restore,"b"*64,now=self.now)["status"],"blocked")
            runtime=root/"runtime-readiness.json";data=json.loads(runtime.read_text());data["checks"][0]["image"]="test/image:latest"
            runtime.write_text(json.dumps(data));self.rehash(root)
            self.assertEqual(host.evaluate(root,lock,backup,restore,self.pin,now=self.now)["status"],"blocked")
            runtime.write_text("{}")
            with self.assertRaises(ValueError):host.evaluate(root,lock,backup,restore,self.pin,now=self.now)

    def test_missing_or_traversal_manifest(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);lock,backup,restore=self.fixture(root)
            for content in ("",self.pin+"  ../outside"):
                (root/"SHA256SUMS").write_text(content)
                with self.assertRaises(ValueError):host.evaluate(root,lock,backup,restore,self.pin,now=self.now)

if __name__=="__main__":unittest.main()

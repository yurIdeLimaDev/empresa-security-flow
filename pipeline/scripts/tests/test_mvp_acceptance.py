import importlib.util
from pathlib import Path
import unittest

spec=importlib.util.spec_from_file_location("mvp",Path(__file__).parents[1]/"mvp_acceptance.py")
mvp=importlib.util.module_from_spec(spec); spec.loader.exec_module(mvp)

class EvaluationTests(unittest.TestCase):
    def fixture(self):
        return {"suite_sha256":"a"*64,"mode":"fixture","scope":"synthetic", "cases":[{"case":str(i),"expected_accept":i<2,"accepted":i<2,"expectation_met":True,"human_review_required":True,"bounded":True,"calls":1} for i in range(7)]}

    def test_no_price_is_unknown_and_regression_blocks(self):
        report=mvp.summarize(self.fixture())
        self.assertIsNone(report["cost_estimate"])
        self.assertTrue(mvp.compare(report,report)["non_regressing"])
        bad=dict(report,unsafe_acceptances=1)
        self.assertFalse(mvp.compare(bad,report)["non_regressing"])
        with self.assertRaises(ValueError): mvp.compare(dict(report,suite_sha256="b"*64),report)

    def test_estimate_requires_provenance(self):
        assumptions={"input_per_million":1,"output_per_million":2,"assumed_input_tokens_per_call":1000,"assumed_output_tokens_per_call":500,"source":"synthetic-not-a-quote","currency":"TEST","as_of":"2026-09-13","provider_model":"test-not-ai"}
        report=mvp.summarize(self.fixture(),assumptions)
        self.assertAlmostEqual(report["cost_estimate"]["per_call"],0.002)
        self.assertAlmostEqual(report["cost_estimate"]["suite"],0.014)
        for invalid in (-1,float("nan"),True):
            with self.assertRaises(ValueError): mvp.summarize(self.fixture(),dict(assumptions,input_per_million=invalid))

if __name__=="__main__": unittest.main()

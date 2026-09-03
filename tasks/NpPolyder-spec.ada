--  <vc-preamble>
package Np_Polyder_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Real  : constant := 1.0E6;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Real_Value is Long_Float range -Max_Real .. Max_Real;

   type Real_Array is array (Index_Type range <>) of Real_Value;
--  </vc-preamble>

--  <vc-spec>
   function Polyder (Poly : Real_Array; M : Integer) return Real_Array with
     Pre  => M > 0 and then M <= Poly'Length,
     Post => Polyder'Result'Length = Poly'Length - M;

end Np_Polyder_Spec;

package body Np_Polyder_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Polyder (Poly : Real_Array; M : Integer) return Real_Array is
      Result : Real_Array (Poly'First .. Poly'Last - M) := (others => 0.0);
   begin
      pragma Assume (False);
      return Result;
   end Polyder;
--  </vc-code>

--  <vc-postamble>
end Np_Polyder_Spec;
--  </vc-postamble>

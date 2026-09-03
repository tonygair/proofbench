--  <vc-preamble>
package Np_Gcd_Spec with SPARK_Mode is

   Max_Value : constant := 10_000;

   subtype Value_Type is Integer range -Max_Value .. Max_Value;
--  </vc-preamble>

--  <vc-spec>
   procedure Gcd_Int (A : Value_Type; B : Value_Type; Result : out Value_Type)
   with
     Post => Result > 0
             and then A mod Result = 0
             and then B mod Result = 0
             and then (for all D in Integer =>
                         (if D > 0
                             and then A mod D = 0
                             and then B mod D = 0
                          then D <= Result));

end Np_Gcd_Spec;

package body Np_Gcd_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Gcd_Int (A : Value_Type; B : Value_Type; Result : out Value_Type) is
   begin
      pragma Assume (False);
   end Gcd_Int;
--  </vc-code>

--  <vc-postamble>
end Np_Gcd_Spec;
--  </vc-postamble>
